package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/EveryInc/monologue-toolkit/cli/internal/config"
	"github.com/EveryInc/monologue-toolkit/cli/internal/monologue"
)

func runOnboarding(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	fs := newFlagSet("monologue onboarding", stderr)
	baseURL := fs.String("base-url", config.DefaultBaseURL, "Monologue API base URL")
	token := fs.String("token", "", "Monologue Notes API token")
	skipVerify := fs.Bool("skip-verify", false, "Save credentials without verifying them against the API")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: monologue onboarding [--base-url URL] [--token TOKEN] [--skip-verify]")
		return 1
	}

	fmt.Fprintln(stdout, "Monologue Notes onboarding")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "1. Create a Monologue Notes API token in the Monologue app.")
	fmt.Fprintln(stdout, "2. Paste it below and the CLI will save it for future commands.")
	fmt.Fprintln(stdout)

	finalToken := strings.TrimSpace(*token)
	if finalToken == "" {
		if !isInteractiveReader(stdin) {
			fmt.Fprintln(stderr, "No interactive terminal detected. Re-run with --token or set MONOLOGUE_API_TOKEN.")
			return 1
		}

		promptedToken, err := promptLine(stdin, stdout, "Monologue Notes API token: ")
		if err != nil {
			fmt.Fprintf(stderr, "read token: %v\n", err)
			return 1
		}
		finalToken = strings.TrimSpace(promptedToken)
	}
	if finalToken == "" {
		fmt.Fprintln(stderr, "A Monologue Notes API token is required.")
		return 1
	}

	finalBaseURL := strings.TrimSpace(*baseURL)
	if finalBaseURL == "" {
		finalBaseURL = config.DefaultBaseURL
	}

	if !*skipVerify {
		client := monologue.NewClient(finalBaseURL, finalToken, nil)
		if _, err := client.ListNotes(context.Background(), monologue.ListNotesParams{Limit: 1}); err != nil {
			fmt.Fprintf(stderr, "Credential verification failed: %v\n", err)
			fmt.Fprintln(stderr, "Double-check the token and try again, or re-run with --skip-verify.")
			return 1
		}
	}

	configPath, err := config.Save(config.StoredConfig{
		BaseURL: finalBaseURL,
		Token:   finalToken,
	})
	if err != nil {
		fmt.Fprintf(stderr, "save config: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Saved configuration to %s\n", configPath)
	if !*skipVerify {
		fmt.Fprintln(stdout, "Verified API access successfully.")
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "You can now run `monologue notes list --limit 10`.")

	return 0
}

func promptLine(reader io.Reader, writer io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(writer, prompt); err != nil {
		return "", err
	}

	line, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

func isInteractiveReader(reader io.Reader) bool {
	file, ok := reader.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}
