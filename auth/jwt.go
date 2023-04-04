package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"net/http"

	"github.com/Frontman-Labs/frontman/config"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type JWTValidator struct {
	issuer   string
	audience string
	JWKS     jwk.Set
}

type JWTValidatorOption func(*JWTValidator)

const AuthTypeBearer string = "bearer"

var (
	ErrMissingAuthHeader   = errors.New("missing authorization header")
	ErrBadFormatAuthHeader = errors.New("invalid format for authorization header")
)

func NewJWTValidator(cfg *config.JWTConfig, opts ...JWTValidatorOption) (*JWTValidator, error) {
	jwks := jwk.NewSet()
	if cfg.KeysUrl != "" {
		keySet, err := jwk.Fetch(context.Background(), cfg.KeysUrl)
		if err != nil {
			log.Printf("Error loading jwks from %s: %s", cfg.KeysUrl, err.Error())
			return nil, err
		}
		jwks = keySet
	}
	validator := &JWTValidator{
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
		JWKS:     jwks,
	}
	for _, opt := range opts {
		opt(validator)
	}
	return validator, nil
}

func (v JWTValidator) ValidateToken(request *http.Request) (map[string]interface{}, error) {
	tokenString := request.Header.Get("Authorization")
	if len(tokenString) == 0 {
		return nil, ErrMissingAuthHeader
	}
	splitToken := strings.Fields(tokenString)
	if len(splitToken) < 2 {
		return nil, ErrBadFormatAuthHeader
	}
	if strings.ToLower(splitToken[0]) != AuthTypeBearer {
		return nil, fmt.Errorf("unsupported authorization type, expected 'Bearer' %w", http.ErrNotSupported)
	}
	token := splitToken[len(splitToken)-1]
	result, err := jwt.Parse([]byte(token), jwt.WithKeySet(v.JWKS, jws.WithInferAlgorithmFromKey(true)))
	if err != nil {
		return nil, err
	}
	fmt.Println(result.Expiration())
	return result.PrivateClaims(), nil
}
