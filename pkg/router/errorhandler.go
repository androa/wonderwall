package router

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nais/wonderwall/pkg/request"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/httplog"
)

type ErrorPage struct {
	CorrelationID        string
	CanonicalRedirectURL string
}

func (h *Handler) respondError(w http.ResponseWriter, r *http.Request, statusCode int, cause error) {
	logger := httplog.LogEntry(r.Context())
	logger.Error().Stack().Err(cause).Msgf("error in route: %+v", cause)

	if len(h.Config.ErrorRedirectURI) > 0 {
		err := h.customErrorRedirect(w, r, statusCode)
		if err == nil {
			return
		}
	}

	h.defaultErrorResponse(w, r, statusCode)
}

func (h *Handler) defaultErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int) {
	w.WriteHeader(statusCode)
	t, _ := template.ParseFiles("templates/error.html")
	errorPage := ErrorPage{
		CorrelationID:        middleware.GetReqID(r.Context()),
		CanonicalRedirectURL: request.CanonicalRedirectURL(r),
	}
	t.Execute(w, errorPage)
}

func (h *Handler) customErrorRedirect(w http.ResponseWriter, r *http.Request, statusCode int) error {
	override, err := url.Parse(h.Config.ErrorRedirectURI)
	if err != nil {
		return err
	}
	// strip scheme and host to avoid cross-domain redirects
	override.Scheme = ""
	override.Host = ""

	query := override.Query()
	query.Add("correlation_id", middleware.GetReqID(r.Context()))
	query.Add("status_code", strconv.Itoa(statusCode))

	override.RawQuery = query.Encode()

	errorRedirectURI := override.String()
	http.Redirect(w, r, errorRedirectURI, http.StatusFound)
	return nil
}

func (h *Handler) InternalError(w http.ResponseWriter, r *http.Request, cause error) {
	h.respondError(w, r, http.StatusInternalServerError, cause)
}

func (h *Handler) BadRequest(w http.ResponseWriter, r *http.Request, cause error) {
	h.respondError(w, r, http.StatusBadRequest, cause)
}

func (h *Handler) Unauthorized(w http.ResponseWriter, r *http.Request, cause error) {
	h.respondError(w, r, http.StatusUnauthorized, cause)
}
