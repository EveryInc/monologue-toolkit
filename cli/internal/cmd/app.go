package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/EveryInc/monologue-toolkit/cli/internal/config"
	"github.com/EveryInc/monologue-toolkit/cli/internal/monologue"
	cliupdate "github.com/EveryInc/monologue-toolkit/cli/internal/update"
	"github.com/EveryInc/monologue-toolkit/cli/internal/version"
)

var (
	checkForUpdate = cliupdate.CheckForUpdate
	updateCLI      = cliupdate.Update
)

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "update" {
		maybeWarnAboutUpdate(stderr)
	}
	if len(args) == 0 {
		printRootUsage(stderr)
		return 1
	}

	switch args[0] {
	case "version", "--version":
		fmt.Fprintln(stdout, version.String())
		return 0
	case "onboarding":
		return runOnboarding(args[1:], stdin, stdout, stderr)
	case "update":
		return runUpdate(args[1:], stdout, stderr)
	case "notes":
		return runNotes(args[1:], stdin, stdout, stderr)
	case "-h", "--help", "help":
		printRootUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printRootUsage(stderr)
		return 1
	}
}

func maybeWarnAboutUpdate(stderr io.Writer) {
	currentVersion := version.Current()
	cachePath, err := config.UpdateCheckPath()
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	latestVersion, outdated, err := checkForUpdate(ctx, currentVersion, cachePath)
	if err != nil || !outdated {
		return
	}
	fmt.Fprintf(stderr, "A newer Monologue CLI version is available (%s; you are using %s). Run `monologue update` to update.\n", latestVersion, currentVersion)
}

func runUpdate(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := newFlagSet("monologue update", stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: monologue update")
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	result, err := updateCLI(ctx, version.Current())
	if err != nil {
		fmt.Fprintf(stderr, "update failed: %v\n", err)
		return 1
	}
	if !result.Updated {
		fmt.Fprintf(stdout, "Monologue CLI is already up to date (%s).\n", result.CurrentVersion)
		return 0
	}
	if result.PendingRestart {
		fmt.Fprintf(stdout, "Monologue CLI %s is downloaded and will finish updating when this command exits.\n", result.LatestVersion)
		return 0
	}
	fmt.Fprintf(stdout, "Updated Monologue CLI from %s to %s.\n", result.CurrentVersion, result.LatestVersion)
	return 0
}

func runNotes(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printNotesUsage(stderr)
		return 1
	}

	switch args[0] {
	case "onboarding":
		return runOnboarding(args[1:], stdin, stdout, stderr)
	case "list":
		return runNotesList(args[1:], stdout, stderr)
	case "all":
		return runNotesAll(args[1:], stdout, stderr)
	case "get":
		return runNotesGet(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printNotesUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown notes command: %s\n\n", args[0])
		printNotesUsage(stderr)
		return 1
	}
}

func runNotesList(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := newFlagSet("monologue notes list", stderr)
	baseURL := fs.String("base-url", "", "Monologue API base URL")
	token := fs.String("token", "", "Monologue Notes API token")
	limit := fs.Int("limit", 20, "Maximum number of notes to return")
	cursor := fs.String("cursor", "", "Opaque pagination cursor")
	query := fs.String("q", "", "Search query across titles, summaries, and transcripts")
	var tagIDs []string
	fs.Func("tag-id", "Filter by tag UUID; repeat to include multiple tags", func(value string) error {
		tagIDs = append(tagIDs, value)
		return nil
	})
	createdAfter := fs.String("created-after", "", "Filter notes created after this ISO 8601 timestamp")
	createdBefore := fs.String("created-before", "", "Filter notes created before this ISO 8601 timestamp")
	updatedAfter := fs.String("updated-after", "", "Filter notes updated after this ISO 8601 timestamp")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	client, ok := newClient(*baseURL, *token, stderr)
	if !ok {
		return 1
	}

	response, err := client.ListNotes(context.Background(), monologue.ListNotesParams{
		Limit:         *limit,
		Cursor:        *cursor,
		Query:         *query,
		TagIDs:        tagIDs,
		CreatedAfter:  *createdAfter,
		CreatedBefore: *createdBefore,
		UpdatedAfter:  *updatedAfter,
	})
	if err != nil {
		return writeError(stderr, err)
	}

	if err := writePrettyJSON(stdout, response); err != nil {
		fmt.Fprintf(stderr, "write response: %v\n", err)
		return 1
	}

	return 0
}

func runNotesAll(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := newFlagSet("monologue notes all", stderr)
	baseURL := fs.String("base-url", "", "Monologue API base URL")
	token := fs.String("token", "", "Monologue Notes API token")
	limit := fs.Int("limit", 100, "Maximum number of notes to request per page")
	cursor := fs.String("cursor", "", "Opaque pagination cursor to resume from")
	query := fs.String("q", "", "Search query across titles, summaries, and transcripts")
	var tagIDs []string
	fs.Func("tag-id", "Filter by tag UUID; repeat to include multiple tags", func(value string) error {
		tagIDs = append(tagIDs, value)
		return nil
	})
	createdAfter := fs.String("created-after", "", "Filter notes created after this ISO 8601 timestamp")
	createdBefore := fs.String("created-before", "", "Filter notes created before this ISO 8601 timestamp")
	updatedAfter := fs.String("updated-after", "", "Filter notes updated after this ISO 8601 timestamp")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	client, ok := newClient(*baseURL, *token, stderr)
	if !ok {
		return 1
	}

	response, err := client.ListAllNotes(context.Background(), monologue.ListNotesParams{
		Limit:         *limit,
		Cursor:        *cursor,
		Query:         *query,
		TagIDs:        tagIDs,
		CreatedAfter:  *createdAfter,
		CreatedBefore: *createdBefore,
		UpdatedAfter:  *updatedAfter,
	})
	if err != nil {
		return writeError(stderr, err)
	}

	if err := writePrettyJSON(stdout, response); err != nil {
		fmt.Fprintf(stderr, "write response: %v\n", err)
		return 1
	}

	return 0
}

func runNotesGet(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := newFlagSet("monologue notes get", stderr)
	baseURL := fs.String("base-url", "", "Monologue API base URL")
	token := fs.String("token", "", "Monologue Notes API token")
	field := fs.String("field", "", "Optional top-level or dotted JSON field to extract")

	// The standard flag package stops parsing at the first positional argument.
	// Accept the documented command in either natural ordering:
	//
	//   monologue notes get NOTE_ID --field transcript
	//   monologue notes get --field transcript NOTE_ID
	//
	noteID := ""
	flagArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		noteID = args[0]
		flagArgs = args[1:]
	}
	if err := fs.Parse(flagArgs); err != nil {
		return 2
	}

	if noteID == "" && fs.NArg() == 1 {
		noteID = fs.Arg(0)
	} else if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: monologue notes get NOTE_ID [--field transcript]")
		return 1
	}
	if noteID == "" {
		fmt.Fprintln(stderr, "usage: monologue notes get NOTE_ID [--field transcript]")
		return 1
	}

	client, ok := newClient(*baseURL, *token, stderr)
	if !ok {
		return 1
	}

	note, err := client.GetNote(context.Background(), noteID)
	if err != nil {
		return writeError(stderr, err)
	}

	if *field == "" {
		if err := writePrettyJSON(stdout, note); err != nil {
			fmt.Fprintf(stderr, "write response: %v\n", err)
			return 1
		}
		return 0
	}

	value, ok := extractJSONPath(note, *field)
	if !ok {
		fmt.Fprintf(stderr, "field not found: %s\n", *field)
		return 1
	}

	if err := writeValue(stdout, value); err != nil {
		fmt.Fprintf(stderr, "write field value: %v\n", err)
		return 1
	}

	return 0
}

func newClient(baseURLFlag string, tokenFlag string, stderr io.Writer) (*monologue.Client, bool) {
	cfg, err := config.Load(baseURLFlag, tokenFlag)
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", err)
		return nil, false
	}
	if cfg.Token == "" {
		fmt.Fprintln(stderr, "No Monologue API token found. Run `monologue onboarding` in your terminal.")
		return nil, false
	}

	return monologue.NewClient(cfg.BaseURL, cfg.Token, nil), true
}

func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

func writePrettyJSON(writer io.Writer, value interface{}) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func writeValue(writer io.Writer, value interface{}) error {
	switch typed := value.(type) {
	case string:
		_, err := fmt.Fprintln(writer, typed)
		return err
	default:
		return writePrettyJSON(writer, typed)
	}
}

func writeError(stderr io.Writer, err error) int {
	if apiError, ok := err.(*monologue.APIError); ok {
		fmt.Fprintln(stderr, apiError.Error())
		return 1
	}

	fmt.Fprintf(stderr, "request failed: %v\n", err)
	return 1
}

func extractJSONPath(value interface{}, path string) (interface{}, bool) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, false
	}

	var current interface{}
	if err := json.Unmarshal(encoded, &current); err != nil {
		return nil, false
	}

	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}

		next, exists := object[part]
		if !exists {
			return nil, false
		}

		current = next
	}

	return current, true
}

func printRootUsage(writer io.Writer) {
	fmt.Fprint(writer, `monologue-toolkit CLI

Usage:
  monologue onboarding [flags]
  monologue update
  monologue notes <command> [flags]

Commands:
  version          Show the installed CLI version
  update           Update the CLI to the latest release
  onboarding       Save and verify Monologue API credentials
  notes onboarding Alias for onboarding
  notes list       List one page of notes
  notes all        Fetch all matching notes across pagination
  notes get        Fetch one note by id

Environment:
  MONOLOGUE_API_TOKEN
  MONOLOGUE_API_BASE_URL
  MONOLOGUE_CONFIG_DIR
`)
}

func printNotesUsage(writer io.Writer) {
	fmt.Fprint(writer, `Usage:
  monologue notes onboarding [flags]
  monologue notes list [flags]
  monologue notes all [flags]
  monologue notes get NOTE_ID [--field transcript]
`)
}
