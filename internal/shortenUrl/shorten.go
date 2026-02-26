package shortenUrl

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/isaacthajunior/url-shortener/internal/sendJson"
	"golang.org/x/net/publicsuffix"
)

type Param struct {
	Url string `json:"url"`
}

func HandleShorten(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var url Param
	err := json.NewDecoder(r.Body).Decode(&url)
	if err != nil {
		sendJson.RespondWithError(w, http.StatusBadRequest, "Err decoding request", err)
		return
	}
	defer r.Body.Close()

	urlString, err := normalizeAndValidateUrl(url.Url)
	if err != nil {
		sendJson.RespondWithError(w, http.StatusBadRequest, "Bad Request", err)
		return
	}

	shortCode, err := shortCodeGenerator(7)
	if err != nil {
		sendJson.RespondWithError(w, http.StatusInternalServerError, "An error occured", err)
		return
	}
	fmt.Println(urlString)
	fmt.Println(shortCode)

}

func normalizeAndValidateUrl(rawString string) (string, error) {
	// 1. Basic cleaning
	u := strings.TrimSpace(rawString)
	if u == "" {
		return "", fmt.Errorf("empty URL")
	}

	// 2. Add https:// if missing
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "https://" + u
	}

	parsedUrl, err := url.ParseRequestURI(u)
	if err != nil {
		return "", err
	}

	// 3. Validate the domain has a valid TLD using publicsuffix
	if err := validateDomainTLD(parsedUrl.Host); err != nil {
		return "", err
	}

	return parsedUrl.String(), nil
}

func validateDomainTLD(host string) error {
	// Remove port if present
	host = strings.Split(host, ":")[0]

	// Check if it's an IP address (skip TLD validation for IPs)
	if isIPAddress(host) {
		return nil // IP addresses are valid even without TLD
	}

	// Must have at least one dot for a domain
	if !strings.Contains(host, ".") {
		return fmt.Errorf("domain must have a TLD (e.g., .com, .org)")
	}

	// Get the public suffix
	suffix, icann := publicsuffix.PublicSuffix(host)

	if suffix == "" {
		return fmt.Errorf("invalid or missing domain extension")
	}

	// Option A: Require ICANN-managed TLDs only
	if !icann {
		return fmt.Errorf("domain extension '%s' is not a standard TLD", suffix)
	}

	return nil
}

// Helper function to check if host is an IP address
func isIPAddress(host string) bool {
	// Simple check - if it consists of numbers and dots, likely an IP
	for part := range strings.SplitSeq(host, ".") {
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return strings.Count(host, ".") == 3 // IPv4 has 3 dots
}

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func shortCodeGenerator(n int) (string, error) {
	randomBytes := make([]byte, n)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	for i := range n {
		randomBytes[i] = base62Chars[int(randomBytes[i])%len(base62Chars)]
	}

	return string(randomBytes), nil
}
