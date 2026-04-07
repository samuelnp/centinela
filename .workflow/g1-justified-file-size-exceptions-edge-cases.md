# Edge Cases: g1-justified-file-size-exceptions

## Covered

- Exception exists with `max_lines` greater than 130 and config load fails.
- Exception path uses backslashes and is normalized for matching.
- Oversized file with valid exception passes and is reported as justified.
- Oversized file with exception but above `max_lines` fails with explicit detail.

## Residual Risks

- Duplicate exception entries for the same path keep the last one.
