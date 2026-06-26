# RoundTable — Agent Instructions

Read **`.cursorrules`** and **`docs/CONSTITUTION.md`** before architectural or implementation work.

When creating git commits, read **`docs/COMMITS.md`** and use the structured commit format.

When running Go tooling for this repo (`go test`, `go mod tidy`, etc.), use:

```bash
export GOPROXY=https://goproxy.cn,direct
```

Or prefer **`make test`** / **`make tidy`** from the repo root (Makefile sets `GOPROXY` automatically).
