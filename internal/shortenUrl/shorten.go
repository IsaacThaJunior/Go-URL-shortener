package shortenUrl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/isaacthajunior/url-shortener/internal/sendJson"
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

	fmt.Println(urlString)

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
	return parsedUrl.String(), nil
}
