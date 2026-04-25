# Session Summary (bl-bernstein-precompute-2.11.19)

## Context
- Goal: integrate Bernstein–Lange precompute tables (GPU) with Kangaroo solver per paper. Legacy kangaroo (no table) must remain correct.
- Environment: GPU RTX 2080 Ti. BL tables in `tables/`. Branch: bl-bernstein-precompute-2.11.19.

## Key Changes
- Disabled default auto-load of BL table; only loads when `--bl-table` is provided.
- Fixed compressed header for BL tables (alpha, walkLength, range info now correct).
- Deterministic BL jump seed; jump generation uses 64-bit uniform steps (but still too small in practice).
- BL path: legacy jump table skipped, DP floor from table, hash collisions enabled (no more discarding non-hint DP).

## Current Status
- Legacy run (no `--bl-table`) on puzzle45 succeeds in ~24s (GPU+CPU), proving core math intact.
- BL runs on puzzle45 still fail to hit; Dead ~400–800, hash memory grows, jump distances only ~1e8 (~2^27), likely too small vs expected ~2^32.6.
- Tables regenerated: puzzle45.bin/txt (26008 entries, dpBits=14, rangePower=44, alpha=0.786, walkLength≈20442). puzzle69 unchanged.

## Known Issues
- BL jump stepSpan may be miscomputed/truncated; runtime jump distances far below L/(4W).
- Need true legacy/BL separation for testing: now OK via `--bl-table` flag.
- BL still no hits; hash pressure high.

## Next Steps
1) Revisit stepSpan calculation (use 128-bit L/(4W)), ensure jumps average ~2^32 for puzzle45; regenerate table.
2) Verify device jump table after upload matches precompute (read back d_blJumpDist) and rerun BL puzzle45.
3) If hits appear, replicate for puzzle69; otherwise inspect DP/adaptive logic.

## Useful Commands
- Legacy test: `./build/kangaroo -gpu -gpuId 0 -g 136,128 config/puzzle45.cfg`
- BL test: `./build/kangaroo -gpu -gpuId 0 -g 136,128 --bl-table tables/puzzle45.bin config/puzzle45.cfg`
- Precompute: `./build/kangaroo_bl_precompute --bl-table tables/puzzle45.txt --bl-start 100000000000 --bl-end 1fffffffffff --gpuId 0 --grid 64 64`
- Logs: `logs/p45_legacy.log`, `logs/p45_bl.log`, `logs/precompute_p45.log`
