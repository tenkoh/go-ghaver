# go-ghaver

A CLI tool that helps you quickly find and select versions of popular GitHub Actions. Search registered actions with fuzzy matching, browse version tags based on releases, and output the selected version (or SHA) for easy use in your workflows.

## Install

```
go install github.com/tenkoh/go-ghaver/cmd/ghaver@latest
```

Run `ghaver` to open the fuzzy finder, search the registered actions, then press Enter to fetch the newest release tags and pick one. The tag prints to stdout with a trailing newline when used interactively and without one when piped. Add `--sha` to print `COMMIT_SHA # TAG` instead of only the tag.

## License
MIT

## Author
tenkoh