# Kangaroo 性能与正确性修复计划（2025-03-02）

目标：消除死亡袋鼠、恢复/提升性能，完成高风险技术债务项。遵守 Cage_Regulation.md，先确保正确性，再优化性能。

## 优先级 P0（立即）
1. **统一 GPU 数据布局**
   - 回滚或彻底修正 SoA：Load/StoreKangaroos、Load/StoreDists、StoreKangaroo、SetKangaroo、主机打包/解包使用同一布局。
   - 移除 limbStride 重复定义，全部采用 limbStride=GPU_GRP_SIZE*laneStride。
   - SetKangaroo 使用 compute/staging stream，同步确保写回完成。
   - 长跑验证 puzzle69（2/20线程），Dead=0。
2. **KSIZE 定义一致**
   - 将 GPU/CPU 的 KSIZE 宏统一（考虑 USE_SYMMETRY），消除潜在越界。
3. **哈希表并发安全与性能**
   - 引入分片锁 + 有序容器（B树/跳表桶）或 lock-free 链表 + 安全回收，保持有序语义，降低 O(n) 插入。
   - 增加锁等待/桶长度指标采样。
4. **DP 队列**
   - 替换 mutex/condvar 为 lock-free 队列或批量出队，减少阻塞；DP 阈值与队列限联动。

## 优先级 P1（性能恢复/提升）
1. **持久化内核完善**
   - 完成工作队列/拷贝同步/多槽协调，默认仍可关闭；提供稳定回退。
2. **模逆/点运算优化**
   - 保留 `_ModInvGrouped`，评估 FMA/融合；CPU 侧可加 AVX2/AVX-512（IntMod）。
3. **DP 计数优化**
   - 在高 DP 率下批量写回，进一步减少原子。
4. **性能基准自动化**
   - 固定基准：puzzle45/puzzle69（2线程、20线程），记录 MK/s、Dead、Lost、GPU/PCIe 利用率；统一日志目录 `logs/`.

## 执行顺序
1) 完成 P0.1/P0.2：统一布局与同步，修复 Dead，长跑验证。  
2) P0.3：哈希表并发改造；P0.4：DP 队列 lock-free。  
3) P1 持久化完善与算术/DP 优化。  
4) 建立性能基准自动化，防止回退。

## 验证与日志
- 每次测试前 kill -9 清理残留进程，日志写入 `logs/`.
- 验证 Dead=0、性能恢复/提升；记录 MK/s、利用率、丢失计数。
