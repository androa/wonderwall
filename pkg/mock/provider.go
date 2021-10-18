package mock

import (
	"github.com/lestrrat-go/jwx/jwk"
	log "github.com/sirupsen/logrus"

	"github.com/nais/wonderwall/pkg/jwks"
	"github.com/nais/wonderwall/pkg/openid"
)

type TestProvider struct {
	ClientConfiguration *TestClientConfiguration
	OpenIDConfiguration *openid.Configuration
	JwksPair            *jwks.Pair
}

func (p TestProvider) GetClientConfiguration() openid.ClientConfiguration {
	return p.ClientConfiguration
}

func (p TestProvider) GetOpenIDConfiguration() *openid.Configuration {
	return p.OpenIDConfiguration
}

func (p TestProvider) GetPublicJwkSet() *jwk.Set {
	return &p.JwksPair.Public
}

func (p TestProvider) PrivateJwkSet() *jwk.Set {
	return &p.JwksPair.Private
}

func NewTestProvider() TestProvider {
	jwksPair, err := jwks.NewJwksPair()
	if err != nil {
		log.Fatal(err)
	}

	clientCfg := clientConfiguration()
	provider := TestProvider{
		ClientConfiguration: &clientCfg,
		OpenIDConfiguration: &openid.Configuration{
			ACRValuesSupported: openid.Supported{"Level3", "Level4"},
			UILocalesSupported: openid.Supported{"nb", "nb", "en", "se"},
		},
		JwksPair: jwksPair,
	}

	return provider
}