package shortenUrl

import (
	"net/http"

	"github.com/isaacthajunior/url-shortener/internal/sendJson"
)

func (h *Handler) Getcompletelink(w http.ResponseWriter, r *http.Request) {
	vars := r.PathValue("code")

	if vars == "" {
		sendJson.RespondWithError(w, http.StatusBadRequest, "No code sent", nil)
		return
	}

	url, err := h.DB.GetUrlByShortCode(r.Context(), vars)
	if err != nil {
		sendJson.RespondWithError(w, http.StatusInternalServerError, "An error occured", err)
		return
	}

	sendJson.RespondWithJSON(w, http.StatusOK, map[string]string{
		"original_url": url.OriginalUrl,
	})

	http.Redirect(w, r, url.OriginalUrl, http.StatusFound)
}
