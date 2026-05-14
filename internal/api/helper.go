package api

import "math/rand"

func String(val string) *string {
	return &val
}
func Bool(val bool) *bool {
	return &val
}
func Int(val int) *int {
	return &val
}

func RandomShortId() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	randomID := make([]byte, 8)
	for i := range randomID {
		randomID[i] = charset[rand.Intn(len(charset))]
	}
	return string(randomID)
}
