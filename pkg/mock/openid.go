package mock

import (
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
)

func IdentityProviderServer() (*httptest.Server, TestProvider) {
	provider := NewTestProvider()
	handler := newIdentityProviderHandler(provider)
	router := identityProviderRouter(handler)
	server := httptest.NewServer(router)

	provider.OpenIDConfiguration.Issuer = server.URL
	provider.OpenIDConfiguration.JwksURI = server.URL + "/jwks"
	provider.OpenIDConfiguration.AuthorizationEndpoint = server.URL + "/authorize"
	provider.OpenIDConfiguration.TokenEndpoint = server.URL + "/token"
	provider.OpenIDConfiguration.EndSessionEndpoint = server.URL + "/endsession"

	return server, provider
}

func identityProviderRouter(ip *identityProviderHandler) chi.Router {
	r := chi.NewRouter()
	r.Get("/authorize", ip.Authorize)
	r.Post("/token", ip.Token)
	r.Get("/jwks", ip.Jwks)
	r.Get("/endsession", ip.EndSession)
	return r
}
