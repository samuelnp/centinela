# Plan: G1 Justified File Size Exceptions

1. Extend config with typed file-size exception entries.
2. Validate exception fields (path, kind, reason, max_lines).
3. Apply exceptions in G1 file-size gate evaluation logic.
4. Fail when exceptions exceed configured max or cap of 130.
5. Add tests for pass/fail/validation branches.
6. Update docs and scaffold templates to explain justified exceptions.
