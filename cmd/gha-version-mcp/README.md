# gha-version-mcp

MCP server that exposes GitHub Actions release tags through a single tool.

## Run

```
go run github.com/tenkoh/go-ghaver/cmd/gha-version-mcp@latest
```

## Tool

`select-gha-version` expects `{ "repository": "owner/name" }` and an optional positive `limit` up to 100. It returns the newest release tags as `tags`, and reports `has_error` with `error_info` when fetching fails.
