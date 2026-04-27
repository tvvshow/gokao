package handlers

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/oktetopython/gaokao/services/recommendation-service/pkg/cppbridge"
)

// CacheWarmOptions 缓存预热配置
type CacheWarmOptions struct {
	Enabled        bool
	Async          bool
	RequestTimeout time.Duration
}

// StartRecommendationCacheWarmup 启动推荐缓存预热
func StartRecommendationCacheWarmup(parent context.Context, logger *logrus.Logger, handler *SimpleRecommendationHandler, options CacheWarmOptions) {
	if handler == nil || handler.bridge == nil || handler.cache == nil || !options.Enabled {
		return
	}
	if logger == nil {
		logger = logrus.New()
	}

	run := func() {
		ctx := parent
		if ctx == nil {
			ctx = context.Background()
		}
		if options.RequestTimeout > 0 {
			timeoutCtx, cancel := context.WithTimeout(ctx, options.RequestTimeout)
			defer cancel()
			ctx = timeoutCtx
		}

		requests := defaultWarmupRequests()
		logger.Infof("开始推荐缓存预热: requests=%d", len(requests))
		summary := handler.WarmRecommendationCache(ctx, requests)
		logger.Infof(
			"推荐缓存预热完成: attempted=%d warmed=%d skipped=%d failed=%d",
			summary.Attempted,
			summary.Warmed,
			summary.Skipped,
			summary.Failed,
		)
	}

	if options.Async {
		go run()
		return
	}
	run()
}

func defaultWarmupRequests() []*cppbridge.RecommendationRequest {
	provinces := []string{"河北", "山东", "河南"}
	scores := []int{550, 600, 650}
	requests := make([]*cppbridge.RecommendationRequest, 0, len(provinces)*len(scores))
	for _, province := range provinces {
		for _, score := range scores {
			requests = append(requests, &cppbridge.RecommendationRequest{
				StudentID:          "cache_warmup",
				Name:               "cache-warmup",
				TotalScore:         score,
				Province:           province,
				SubjectCombination: "物理",
				MaxRecommendations: 30,
				Algorithm:          "hybrid",
				Preferences: map[string]interface{}{
					"risk_tolerance": "moderate",
				},
			})
		}
	}
	return requests
}
