package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/EveryInc/monologue-toolkit/cli/internal/config"
	"github.com/EveryInc/monologue-toolkit/cli/internal/monologue"
)

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printRootUsage(stderr)
		return 1
	}

	switch args[0] {
	case "onboarding":
		return runOnboarding(args[1:], stdin, stdout, stderr)
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
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: monologue notes get NOTE_ID [--field transcript]")
		return 1
	}

	client, ok := newClient(*baseURL, *token, stderr)
	if !ok {
		return 1
	}

	note, err := client.GetNote(context.Background(), fs.Arg(0))
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
		fmt.Fprintln(stderr, "No Monologue API token found. Run `monologue onboarding` or pass --token.")
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
  monologue notes <command> [flags]

Commands:
  onboarding       Save and verify Monologue API credentials
  notes onboarding Alias for onboarding
  notes list       List one page of notes
  notes all        Fetch all matching notes across pagination
  notes get        Fetch one note by id

Environment:
  MONOLOGUE_API_TOKEN
  MONOLOGUE_API_BASE_URL
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
