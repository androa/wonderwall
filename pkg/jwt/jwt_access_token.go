package jwt

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type AccessToken struct {
	Token
}

func NewAccessToken(raw string, jwtToken jwt.Token) *AccessToken {
	return &AccessToken{
		NewToken(raw, jwtToken),
	}
}

func ParseAccessToken(raw string, jwks jwk.Set) (*AccessToken, error) {
	accessToken, err := Parse(raw, jwks)
	if err != nil {
		return nil, err
	}

	return NewAccessToken(raw, accessToken), nil
}
