// Package credentials handles loading, storing and deleting Porkbun API credentials.
package credentials

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

const (
	// EnvApiKey and EnvSecretKey are the environment variables that can be used to provide credentials.
	EnvApiKey    = "PORKBUN_API_KEY"
	EnvSecretKey = "PORKBUN_SECRET_KEY"

	appDirName = "porkbun-dns"
	fileName   = "credentials.toml"
)

// Source describes where credentials were found.
type Source string

const (
	SourceNone        Source = ""
	SourceEnvironment Source = "environment"
	SourceFile        Source = "file"
)

// ErrNotFound is returned when no stored credentials exist.
var ErrNotFound = errors.New("no credentials found")

// Credentials are the Porkbun API key and secret key.
type Credentials struct {
	ApiKey    string `toml:"api_key"`
	SecretKey string `toml:"secret_key"`
}

// IsComplete returns true if both the API key and the secret key are set.
func (c Credentials) IsComplete() bool {
	return c.ApiKey != "" && c.SecretKey != ""
}

// FilePath returns the OS-specific path of the credentials file. It is based on os.UserConfigDir, so it resolves to
// $XDG_CONFIG_HOME (or ~/.config) on Linux, ~/Library/Application Support on macOS, and %AppData% on Windows.
func FilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user config directory: %w", err)
	}
	return filepath.Join(dir, appDirName, fileName), nil
}

// FromEnvironment reads the credentials from the PORKBUN_API_KEY and PORKBUN_SECRET_KEY environment variables.
func FromEnvironment() Credentials {
	return Credentials{
		ApiKey:    os.Getenv(EnvApiKey),
		SecretKey: os.Getenv(EnvSecretKey),
	}
}

// Load reads the credentials from the credentials file. It returns ErrNotFound if the file does not exist.
func Load() (Credentials, error) {
	path, err := FilePath()
	if err != nil {
		return Credentials{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Credentials{}, ErrNotFound
		}
		return Credentials{}, err
	}

	var creds Credentials
	if err := toml.Unmarshal(data, &creds); err != nil {
		return Credentials{}, fmt.Errorf("invalid credentials file %s: %w", path, err)
	}
	return creds, nil
}

// Save writes the credentials to the credentials file (unencrypted) with restrictive permissions, and returns the
// path of the file.
func Save(creds Credentials) (string, error) {
	path, err := FilePath()
	if err != nil {
		return "", err
	}

	data, err := toml.Marshal(creds)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	// WriteFile does not change the permissions of an existing file
	if err := os.Chmod(path, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// Delete removes the credentials file. It returns ErrNotFound if there was nothing to delete, and the path of the
// credentials file.
func Delete() (string, error) {
	path, err := FilePath()
	if err != nil {
		return "", err
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return path, ErrNotFound
		}
		return path, err
	}
	return path, nil
}

// Resolve returns the first complete set of credentials, checking the environment first and then the credentials
// file. If none are found, the returned source is SourceNone and the error is nil.
func Resolve() (Credentials, Source, error) {
	if creds := FromEnvironment(); creds.IsComplete() {
		return creds, SourceEnvironment, nil
	}

	creds, err := Load()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Credentials{}, SourceNone, nil
		}
		return Credentials{}, SourceNone, err
	}
	if creds.IsComplete() {
		return creds, SourceFile, nil
	}
	return Credentials{}, SourceNone, nil
}
