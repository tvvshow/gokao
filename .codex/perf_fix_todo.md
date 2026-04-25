# 计算性能极致优化 TODO（精准执行版）

优先级仅围绕“极致速度+正确性”，避免过度工程。

1) **KSIZE 一致性防护**
   - 在 GPU/CPU 公共头（如 Constants/Types）加 static_assert/constexpr，USE_SYMMETRY=11，否则 10。
2) **单点更新同步保障**
   - SetKangaroo：沿用 stagingStream + 同步，禁止默认 stream；如可，直接使用 compute stream。
3) **锁争用热点优化（轻量）**
   - HashTable：增加分片锁（按桶片段分组），减少 contention；保持有序数组实现，少改动。
   - DP 队列：简单提升 DP_QUEUE_LIMIT/WORKER_COUNT 已做；进一步可用轻量 SPSC 或批量 pop 减少锁。
4) **核态性能**
   - 保持 AoS，使用 `_ModInvGrouped`；跳跃哈希已上，warp DP 归约保留。
   - 可评估恢复多槽（4）仅在持久化关闭情况下，避免复杂控制。
5) **基准验证**
   - 每次改动后跑 puzzle45/puzzle69（2/20 线程），记录 MK/s、Dead、Lost，日志入 logs/。
