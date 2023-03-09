package auth

import (
	"errors"
	"github.com/Frontman-Labs/frontman/config"
)

type TokenValidator interface {
	ValidateToken(tokenString string) (map[string]interface{}, error)
}

func GetTokenValidator(conf config.AuthConfig) (TokenValidator, error) {
	switch conf.AuthType {
	case "jwt":
		return NewJWTValidator(conf.JWT.Audience, conf.JWT.Issuer, conf.JWT.KeysUrl), nil
	default:
		return nil, errors.New("Unrecognized auth type specified")
	}
}
