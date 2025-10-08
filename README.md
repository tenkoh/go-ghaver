# go-ghaver

`ghaver` is a CLI helper for picking GitHub Actions versions without leaving the terminal.

## Why
- Action names blur together and it is easy to forget the exact `owner/repo`.
- Checking what the newest version is means interrupting your flow to open the Marketplace or releases page.

`ghaver` bundles a curated list of popular actions, lets you fuzzy-find the one you want, and immediately shows the published tags (plus commit SHAs when needed).

## Quick Example
```
$ ghaver
actions/checkout@v4

$ ghaver --sha
actions/checkout@<sha-value> # v4.3.0
```

## Install

```
go install github.com/tenkoh/go-ghaver/cmd/ghaver@latest
```

## Usage
Run `ghaver` to open the fuzzy finder, search the registered actions, then press Enter to fetch the newest release tags and pick one. The selected version prints to stdout as `<owner>/<repo>@<tag>`. Add `--sha` to output `<owner>/<repo>@<sha> # <tag>` instead.

When run interactively a trailing newline is added for convenience; when piped the newline is omitted so you can drop the result straight into workflow files.

## License
MIT

## Author
tenkoh
