<!-- centinela:doc-version=1 template=docs/architecture/stack-checks-reference.md -->
# Stack-Specific Production Readiness Checks — Reference

This file holds the multi-language example matrix that used to live inline
in `production-readiness-prompt.md.template`. Each generated
`production-readiness-prompt.md` only needs the bullets relevant to its
project's stack; everything else is reference material kept here for
maintainers and for new-project bootstrap.

## Examples by language

### Go
- Goroutine leaks: every `go func()` either has a clear exit path or is
  bounded by `context`.
- Missing `defer Close()` on `os.File`, `*sql.Rows`, etc.
- Unhandled errors from secondary calls (`rows.Err()`, `f.Close()` after
  write).
- Context propagation: every external/blocking call accepts and respects
  `ctx`.

### TypeScript
- Unhandled Promise rejections; missing `await` on async calls.
- `any` casts hiding errors; missing `.catch()` on floating promises.
- Resource cleanup with `using` declarations or `finally` blocks where
  AsyncDisposable applies.

### Python
- Bare `except:` clauses without re-raise.
- Missing `async with` for async context managers.
- Unclosed file handles; missing `finally` for cleanup.
- Bounded retry helpers, not while-True loops.

### Ruby
- `rescue` without re-raise on unexpected errors.
- Missing transaction rollback in multi-step writes.
- Missing `ensure` for resource cleanup.

## How this file is referenced

`production-readiness-prompt.md.template` keeps a single inline
placeholder pointing back here. When `centinela init` renders the
template into a project's `production-readiness-prompt.md`, the project's
own stack bullets are inlined and the unrelated languages are left out.
