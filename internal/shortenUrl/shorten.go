package shortenUrl

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/isaacthajunior/url-shortener/internal/database"
	"github.com/isaacthajunior/url-shortener/internal/sendJson"
	"github.com/lib/pq"
	"golang.org/x/net/publicsuffix"
)

type Handler struct {
	DB *database.Queries
}

type Param struct {
	Url string `json:"url"`
}

func (h *Handler) HandleShorten(w http.ResponseWriter, r *http.Request) {
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

	const maxRetries = 3
	const myDomain = "isaac.edu/"

	for range maxRetries {
		shortCode, err := shortCodeGenerator(7)
		if err != nil {
			sendJson.RespondWithError(w, http.StatusInternalServerError, "Failed to generate short code", err)
			return
		}

		_, err = h.DB.CreateUrl(r.Context(), database.CreateUrlParams{
			ID:          uuid.New(),
			OriginalUrl: urlString,
			ShortCode:   shortCode,
		})

		// No error → break out and return success later
		if err == nil {
			sendJson.RespondWithJSON(w, http.StatusOK, map[string]string{
				"short_url": "https://" + myDomain + shortCode,
			})
			return
		}

		pqErr, ok := err.(*pq.Error)
		if !ok {
			sendJson.RespondWithError(w, http.StatusInternalServerError, "Database error", err)
			return
		}

		if pqErr.Code != "23505" {
			sendJson.RespondWithError(w, http.StatusInternalServerError, "Database error", err)
			return
		}

		// Unique violation — determine which constraint
		switch pqErr.Constraint {

		case "urls_short_code_key":
			// collision → retry
			continue

		case "urls_original_url_key":
			// URL already exists → fetch and return

			existing, err := h.DB.GetUrlByOriginalUrl(r.Context(), urlString)
			if err != nil {
				sendJson.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch existing URL", err)
				return
			}

			sendJson.RespondWithJSON(w, http.StatusOK, map[string]string{
				"short_url": "https://" + myDomain + existing.ShortCode,
			})

			log.Println("The URL is not new. We have already generated a code for it before")
			return

		default:
			sendJson.RespondWithError(w, http.StatusInternalServerError, "Database error", err)
			return
		}
	}

	sendJson.RespondWithError(w, http.StatusInternalServerError, "Failed to generate unique short code", nil)

}

func normalizeAndValidateUrl(rawString string) (string, error) {
	// 1. Basic cleaning
	u := strings.TrimSpace(rawString)
	if u == "" {
		return "", fmt.Errorf("empty URL")
	}

	// turn to lowercase
	u = strings.ToLower(u)

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
