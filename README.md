# go-ghaver

Tools for browsing GitHub Actions release tags.

## ghaver CLI

```
go install github.com/tenkoh/go-ghaver/cmd/ghaver@latest
```

Run `ghaver` to open the fuzzy finder, search the registered actions, then press Enter to fetch the newest release tags and pick one. The tag prints to stdout with a trailing newline when used interactively and without one when piped. Add `--sha` to print `COMMIT_SHA # TAG` instead of only the tag.

## gha-version-mcp Server

```
go run github.com/tenkoh/go-ghaver/cmd/gha-version-mcp@latest
```

This starts an MCP server exposing the `select-gha-version` tool. Provide `{ "repository": "owner/name" }` and an optional positive `limit` (max 100). The response includes `tags` for the newest releases, or sets `has_error` with `error_info` when fetching fails.

## License
MIT

## Author
tenkoh