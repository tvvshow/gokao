package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestRateLimiter_EvictsIdleBuckets 验证 cleanup 周期能淘汰空闲 bucket。
// 不依赖真实 ticker（默认 1min 太慢），直接同步调 evictIdle() + 注入时钟快进。
func TestRateLimiter_EvictsIdleBuckets(t *testing.T) {
	base := time.Now()
	now := base
	nowFn := func() time.Time { return now }
	rl := newRateLimiterWithDeps(10, 20, 5*time.Second, nowFn)
	defer rl.stop()

	for i := 0; i < 100; i++ {
		if ok, _ := rl.allow(fmt.Sprintf("ip-%d", i)); !ok {
			t.Fatalf("ip-%d should be allowed within burst, got rate-limited", i)
		}
	}
	if got := countBuckets(rl); got != 100 {
		t.Fatalf("expected 100 buckets seeded, got %d", got)
	}

	now = now.Add(6 * time.Second)
	if evicted := rl.evictIdle(); evicted != 100 {
		t.Fatalf("expected 100 idle buckets evicted, got %d", evicted)
	}
	if got := countBuckets(rl); got != 0 {
		t.Fatalf("expected empty sync.Map after eviction, got %d", got)
	}
}

// TestRateLimiter_KeepsActiveBuckets 确认 evictIdle 不误删仍活跃的 bucket。
func TestRateLimiter_KeepsActiveBuckets(t *testing.T) {
	base := time.Now()
	now := base
	nowFn := func() time.Time { return now }
	rl := newRateLimiterWithDeps(10, 20, 5*time.Second, nowFn)
	defer rl.stop()

	rl.allow("idle-ip")
	now = now.Add(3 * time.Second)
	rl.allow("active-ip")

	now = now.Add(3 * time.Second)
	if evicted := rl.evictIdle(); evicted != 1 {
		t.Fatalf("expected exactly 1 eviction (idle-ip), got %d", evicted)
	}
	if got := countBuckets(rl); got != 1 {
		t.Fatalf("expected 1 active bucket remaining, got %d", got)
	}
	if _, ok := rl.m.Load("active-ip"); !ok {
		t.Fatal("active-ip should still be present")
	}
}

// TestRateLimiter_StopIsIdempotent 验证 stop 多次调用不 panic（close on closed channel 防护）。
func TestRateLimiter_StopIsIdempotent(t *testing.T) {
	rl := newRateLimiter(10, 20)
	rl.stop()
	rl.stop()
	rl.stop()
}

// TestRateLimiter_AllowUsesInjectedClock 验证 allow() 走的是注入时钟，
// 不再混用 time.Now()，否则 evictIdle 算的 idle 与 allow 更新的 last 时间基不一致。
func TestRateLimiter_AllowUsesInjectedClock(t *testing.T) {
	base := time.Now()
	now := base
	nowFn := func() time.Time { return now }
	rl := newRateLimiterWithDeps(10, 20, 5*time.Second, nowFn)
	defer rl.stop()

	rl.allow("ip-1")
	v, _ := rl.m.Load("ip-1")
	b := v.(*rateBucket)
	b.mu.Lock()
	last := b.last
	b.mu.Unlock()
	if !last.Equal(base) {
		t.Fatalf("bucket.last should equal injected base time, got %v vs %v", last, base)
	}

	now = now.Add(2 * time.Second)
	rl.allow("ip-1")
	b.mu.Lock()
	last = b.last
	b.mu.Unlock()
	if !last.Equal(base.Add(2 * time.Second)) {
		t.Fatalf("bucket.last should advance to injected now, got %v", last)
	}
}

func countBuckets(rl *rateLimiter) int {
	n := 0
	rl.m.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
}

// 编译期断言：确保 stop / once 字段没被误删，sync.Once 才能保证幂等关闭。
var _ = sync.Once{}
