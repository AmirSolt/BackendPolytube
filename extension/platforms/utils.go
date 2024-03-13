package platforms

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
)

func buildURLFromMap(baseURL string, queries map[string]string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for key, value := range queries {
		q.Add(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func generateCSRFState() (string, error) {
	// Create a byte slice to hold the random bytes
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Encode the random bytes to hex string (base 16) and remove the "0x" prefix
	return hex.EncodeToString(b)[2:], nil
}
