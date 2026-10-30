package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thrau/porkbun-dyndns/internal/api"
	"github.com/thrau/porkbun-dyndns/internal/credentials"
	"golang.org/x/term"
)

func NewAuthCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Porkbun API credentials",
	}

	cmd.AddCommand(NewAuthSetCommand(app))
	cmd.AddCommand(NewAuthStatusCommand(app))
	cmd.AddCommand(NewAuthDeleteCommand(app))

	return cmd
}

// NewAuthSetCommand creates the interactive command that asks for the API key and secret key.
func NewAuthSetCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Interactively set the Porkbun API key and secret key",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			return runAuthSet(cmd, app)
		},
	}
}

// NewAuthStatusCommand creates the non-interactive command that shows the current authentication status.
func NewAuthStatusCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the current authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			return runAuthStatus(cmd, app)
		},
	}
}

// NewAuthDeleteCommand creates the non-interactive command that deletes the stored credentials.
func NewAuthDeleteCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete the stored Porkbun credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			return runAuthDelete(cmd, app)
		},
	}
}

// runAuthSet prompts the user for the API key and secret key and stores them.
func runAuthSet(cmd *cobra.Command, app *App) error {
	reader := bufio.NewReader(cmd.InOrStdin())
	out := cmd.OutOrStdout()

	apiKey, err := promptValue(reader, out, "API key: ", false)
	if err != nil {
		return err
	}

	secretKey, err := promptValue(reader, out, "Secret key: ", true)
	if err != nil {
		return err
	}

	path, err := credentials.Save(credentials.Credentials{ApiKey: apiKey, SecretKey: secretKey})
	if err != nil {
		return fmt.Errorf("could not store credentials: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Credentials stored in %s\n", path)

	return nil
}

// promptValue prints the prompt and reads a non-empty value. If hidden is true and stdin is a terminal, the input
// is not echoed.
func promptValue(reader *bufio.Reader, out io.Writer, prompt string, hidden bool) (string, error) {
	_, _ = fmt.Fprint(out, prompt)

	var value string
	if fd := int(os.Stdin.Fd()); hidden && term.IsTerminal(fd) {
		raw, err := term.ReadPassword(fd)
		_, _ = fmt.Fprintln(out)
		if err != nil {
			return "", err
		}
		value = string(raw)
	} else {
		line, err := reader.ReadString('\n')
		if err != nil && (err != io.EOF || line == "") {
			return "", err
		}
		value = line
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s must not be empty", strings.TrimSuffix(prompt, ": "))
	}

	return value, nil
}

// runAuthStatus reports where credentials are coming from and whether they are valid.
func runAuthStatus(cmd *cobra.Command, app *App) error {
	out := cmd.OutOrStdout()

	path, err := credentials.FilePath()
	if err != nil {
		return err
	}

	stored, err := credentials.Load()
	switch {
	case errors.Is(err, credentials.ErrNotFound):
		_, _ = fmt.Fprintf(out, "Credentials file: %s (not found)\n", path)
	case err != nil:
		return err
	case stored.IsComplete():
		_, _ = fmt.Fprintf(out, "Credentials file: %s (credentials set)\n", path)
	default:
		_, _ = fmt.Fprintf(out, "Credentials file: %s (incomplete credentials)\n", path)
	}

	envState := "not set"
	if credentials.FromEnvironment().IsComplete() {
		envState = "credentials set"
	}
	_, _ = fmt.Fprintf(out, "Environment (%s, %s): %s\n", credentials.EnvApiKey, credentials.EnvSecretKey, envState)

	creds, source, err := credentials.Resolve()
	if err != nil {
		return err
	}
	if source == credentials.SourceNone {
		return errors.New("no credentials found: run `porkbun-dns auth set` or set " +
			credentials.EnvApiKey + " and " + credentials.EnvSecretKey)
	}
	_, _ = fmt.Fprintf(out, "Active credentials source: %s\n", source)

	client := api.NewClient(api.WithCredentials(creds.ApiKey, creds.SecretKey))
	resp, err := client.Ping(cmd.Context())
	if err != nil {
		_, _ = fmt.Fprintln(out, "Ping: failed")
		return fmt.Errorf("credentials are not valid: %w", err)
	}
	_, _ = fmt.Fprintf(out, "Ping: ok (your IP: %s)\n", resp.YourIp)

	return nil
}

// runAuthDelete removes the stored credentials.
func runAuthDelete(cmd *cobra.Command, app *App) error {
	out := cmd.OutOrStdout()

	path, err := credentials.Delete()
	if err != nil {
		if errors.Is(err, credentials.ErrNotFound) {
			_, _ = fmt.Fprintf(out, "No stored credentials found (%s)\n", path)
			return nil
		}
		return fmt.Errorf("could not delete credentials: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Deleted credentials file %s\n", path)
	if credentials.FromEnvironment().IsComplete() {
		_, _ = fmt.Fprintf(out, "Note: credentials are still set in the environment (%s, %s)\n",
			credentials.EnvApiKey, credentials.EnvSecretKey)
	}

	return nil
}
