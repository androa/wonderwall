package router

import (
	"fmt"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/nais/wonderwall/pkg/session"
	"github.com/nais/wonderwall/pkg/token"
	"golang.org/x/oauth2"
	"net/http"
	"time"
)

// localSessionID prefixes the given `sid` with the given client ID to prevent key collisions.
// `sid` is a key that refers to the user's unique SSO session at the Identity Provider, and the same key is present
// in all tokens acquired by any Relying Party (such as Wonderwall) during that session.
// Thus, we cannot assume that the value of `sid` to uniquely identify the pair of (user, application session)
// if using a shared session store.
func (h *Handler) localSessionID(sid string) string {
	return fmt.Sprintf("%s-%s", h.Config.ClientID, sid)
}

func (h *Handler) getSessionFromCookie(r *http.Request) (*session.Data, error) {
	sessionID, err := h.getEncryptedCookie(r, h.GetSessionCookieName())
	if err != nil {
		return nil, fmt.Errorf("no session cookie: %w", err)
	}

	encryptedSessionData, err := h.Sessions.Read(r.Context(), sessionID)
	if err != nil {
		return nil, fmt.Errorf("reading session from store: %w", err)
	}

	sessionData, err := encryptedSessionData.Decrypt(h.Crypter)
	if err != nil {
		return nil, fmt.Errorf("decrypting session data: %w", err)
	}

	return sessionData, nil
}

func (h *Handler) getSessionLifetime(accessToken string) (time.Duration, error) {
	defaultSessionLifetime := h.Config.SessionMaxLifetime

	token, err := jwt.Parse([]byte(accessToken))
	if err != nil {
		return 0, err
	}

	tokenDuration := token.Expiration().Sub(time.Now())

	if tokenDuration <= defaultSessionLifetime {
		return tokenDuration, nil
	}

	return defaultSessionLifetime, nil
}

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request, externalSessionID string, tokens *oauth2.Token, idToken *token.IDToken) error {
	sessionID := h.localSessionID(externalSessionID)

	sessionLifetime, err := h.getSessionLifetime(tokens.AccessToken)
	if err != nil {
		return fmt.Errorf("getting access token lifetime: %w", err)
	}

	err = h.setEncryptedCookie(w, h.GetSessionCookieName(), sessionID, sessionLifetime)
	if err != nil {
		return fmt.Errorf("setting session cookie: %w", err)
	}

	sessionData := &session.Data{
		ExternalSessionID: externalSessionID,
		OAuth2Token:       tokens,
		IDTokenSerialized: idToken.Raw,
	}

	encryptedSessionData, err := sessionData.Encrypt(h.Crypter)
	if err != nil {
		return fmt.Errorf("encrypting session data: %w", err)
	}

	err = h.Sessions.Write(r.Context(), sessionID, encryptedSessionData, sessionLifetime)
	if err != nil {
		return fmt.Errorf("writing session to store: %w", err)
	}

	return nil
}

func (h *Handler) getAccessTokenFromSession(r *http.Request, sessionID string) (jwt.Token, error) {
	encryptedSession, err := h.Sessions.Read(r.Context(), sessionID)
	if err != nil {
		// Session not found; ignoring
		return nil, nil
	}

	sessionData, err := encryptedSession.Decrypt(h.Crypter)
	if err != nil {
		// Can't decrypt, likely not our session; ignoring
		return nil, nil
	}

	accessToken, err := jwt.Parse([]byte(sessionData.OAuth2Token.AccessToken))
	if err != nil {
		return nil, fmt.Errorf("parsing session access token: %w", err)
	}

	return accessToken, nil
}
