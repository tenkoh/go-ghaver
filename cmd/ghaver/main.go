package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/tenkoh/go-ghaver/internal/ghaver"
)

//go:embed actions.json
var actionsJSON []byte

// version is overridden at build time via -ldflags.
var version = "dev"

type cli struct {
	SHA     bool `kong:"help='Include commit SHA in the output'"`
	Version bool `kong:"help='Print version information and exit',short='v',name='version'"`
}

func main() {
	cfg, parser, err := parseCLI(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}
	if cfg.Version {
		fmt.Fprintln(os.Stdout, version)
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, cfg); err != nil {
		if errors.Is(err, ghaver.ErrSelectionAborted) || errors.Is(err, context.Canceled) {
			return
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		parser.Exit(1)
	}
}

func parseCLI(args []string) (cli, *kong.Kong, error) {
	var cfg cli
	parser := kong.Must(&cfg,
		kong.Name("ghaver"),
		kong.Description("Search and select GitHub Actions versions."),
	)
	if _, err := parser.Parse(args); err != nil {
		return cli{}, parser, err
	}
	return cfg, parser, nil
}

func run(ctx context.Context, cfg cli) error {
	actions, err := ghaver.ActionsFromJSON(actionsJSON)
	if err != nil {
		return err
	}

	selector, err := ghaver.NewFZFSelector()
	if err != nil {
		return err
	}
	defer selector.Close()

	runner := ghaver.Runner{
		Actions:  actions,
		Selector: selector,
		Client:   ghaver.NewGitHubAPI(nil),
	}

	version, err := runner.Run(ctx, ghaver.Options{WithSHA: cfg.SHA})
	if err != nil {
		return err
	}

	repoRef := fmt.Sprintf("%s/%s", version.Owner, version.Repo)
	output := fmt.Sprintf("%s@%s", repoRef, version.Tag)
	if cfg.SHA {
		output = fmt.Sprintf("%s@%s # %s", repoRef, version.SHA, version.Tag)
	}

	if shouldAddNewline(os.Stdout) {
		fmt.Fprintln(os.Stdout, output)
	} else {
		fmt.Fprint(os.Stdout, output)
	}

	return nil
}

func shouldAddNewline(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return true
	}
	return info.Mode()&os.ModeCharDevice != 0
}
