# Legacy vs BL Current State
- Legacy (no --bl-table): puzzle45 solves in ~24s, math chain intact.
- BL (--bl-table): still no hit; Dead ~400–800, hash mem grows.
- Jump distances: ~1e8 (2^27), too small vs expected ~2^32.6.
- Tables: puzzle45 regenerated (26008 entries, dpBits=14, rangePower=44, alpha=0.786, walkLength≈20442). puzzle69 unchanged.
- BL/legacy separation fixed: only loads table when --bl-table is provided.
- Pending: stepSpan fix, jump distribution alignment, regenerate table, retest.
