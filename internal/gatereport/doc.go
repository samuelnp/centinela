// Package gatereport parses the validate-step verifier report,
// `.workflow/<feature>-gatekeeper.md`. It is pure: no filesystem, no git.
//
// Two contracts live here and nowhere else:
//
//   - The verdict contract (PR #59): the verdict is the FIRST token of the
//     `**Status:**` line. The rest of the report is never scanned, so prose
//     mentioning "warnings" can never skew a SAFE verdict.
//   - The verification contract (D2): a report is grounded only when it
//     carries a ```json centinela:verification fenced block recording the
//     revision + tree digest it verified and the commands it actually ran.
//
// Missing or unparseable input is always an error. Nothing here fails open.
package gatereport
