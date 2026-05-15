package util

import (
	"math/rand"
	"strings"
)

// SplitDomain splits a fully qualified domain name into the domain and subdomain portions.
// For example, "www.example.com" would return "example.com", "www"; "foo.bar.example.com" would return "example.com",
// "foo.bar". If the name does not contain a subdomain, the subdomain will be an empty string.
func SplitDomain(name string) (domain string, subdomain string) {
	parts := strings.Split(name, ".")
	domain = name
	subdomain = ""
	if len(parts) > 2 {
		domain = strings.Join(parts[len(parts)-2:], ".")
		subdomain = strings.Join(parts[:len(parts)-2], ".")
	}
	return domain, subdomain
}

// RandomShortId generates an 8-char random identifier consisting of lowercase letters and digits.
func RandomShortId() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	randomID := make([]byte, 8)
	for i := range randomID {
		randomID[i] = charset[rand.Intn(len(charset))]
	}
	return string(randomID)
}
