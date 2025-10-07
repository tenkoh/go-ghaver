package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tenkoh/go-ghaver/internal/ghaver"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type selectInput struct {
	Repository string `json:"repository" jsonschema:"GitHub Action repository in the form owner/name"`
	Limit      *int   `json:"limit,omitempty" jsonschema:"Maximum number of release tags to return (defaults to 20, max 100)"`
}

type selectOutput struct {
	Tags      []string `json:"tags" jsonschema:"Release tags ordered from newest to oldest"`
	HasError  bool     `json:"has_error" jsonschema:"Set to true when fetching release tags failed"`
	ErrorInfo string   `json:"error_info,omitempty" jsonschema:"Details about the failure when has_error is true"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := ghaver.NewGitHubAPI(nil)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "gha-version-mcp",
		Version: "0.1.0",
		Title:   "GitHub Actions Version Select Helper",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "select-gha-version",
		Description: "Fetch release tags for a GitHub Action repository and surface them to the caller. When the newest releases belong to a fresh major version with limited history, explicitly ask the user whether they want the absolute latest release (including that major) or the newest release from the prior major before making a recommendation.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input selectInput) (*mcp.CallToolResult, selectOutput, error) {
		output, err := handleRequest(ctx, client, input)
		if err != nil {
			output.HasError = true
			output.ErrorInfo = err.Error()
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, output, nil
		}
		return nil, output, nil
	})

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func handleRequest(ctx context.Context, client ghaver.GitHubClient, input selectInput) (selectOutput, error) {
	repo := strings.TrimSpace(input.Repository)
	if repo == "" {
		return selectOutput{}, errors.New("repository is required and must be in the form owner/name")
	}

	owner, name, ok := strings.Cut(repo, "/")
	if !ok || owner == "" || name == "" {
		return selectOutput{}, fmt.Errorf("invalid repository %q: expected owner/name", repo)
	}

	limit := defaultLimit
	if input.Limit != nil {
		switch {
		case *input.Limit <= 0:
			return selectOutput{}, fmt.Errorf("limit must be positive, got %d", *input.Limit)
		case *input.Limit > maxLimit:
			limit = maxLimit
		default:
			limit = *input.Limit
		}
	}

	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	versions, err := client.Versions(fetchCtx, owner, name)
	if err != nil {
		return selectOutput{}, err
	}

	tags := make([]string, 0, limit)
	seen := make(map[string]struct{})
	for _, v := range versions {
		if v.Tag == "" {
			continue
		}
		if _, exists := seen[v.Tag]; exists {
			continue
		}
		seen[v.Tag] = struct{}{}
		tags = append(tags, v.Tag)
		if len(tags) >= limit {
			break
		}
	}

	return selectOutput{Tags: tags}, nil
}
