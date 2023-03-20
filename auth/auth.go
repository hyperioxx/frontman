package auth

import (
	"errors"
	"github.com/Frontman-Labs/frontman/config"
	"net/http"
)

type TokenValidator interface {
	ValidateToken(request *http.Request) (map[string]interface{}, error)
}

func GetTokenValidator(conf config.AuthConfig) (TokenValidator, error) {
	switch conf.AuthType {
	case "jwt":
		return NewJWTValidator(conf.JWT.Audience, conf.JWT.Issuer, conf.JWT.KeysUrl)
	case "basic":
		return NewBasicAuthValidator(conf.BasicAuthConfig)
	default:
		return nil, errors.New("Unrecognized auth type specified")
	}
}
