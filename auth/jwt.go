package auth

import (
	"context"
	"fmt"
	"log"
	"strings"

	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type JWTValidator struct {
	issuer   string
	audience string
	JWKS     jwk.Set
}

const AuthTypeBearer string = "bearer"

func NewJWTValidator(issuer string, audience string, jwkUrl string) (*JWTValidator, error) {
	jwks, err := jwk.Fetch(context.Background(), jwkUrl)
	if err != nil {
		log.Printf("Error loading jwks from %s: %s", jwkUrl, err.Error())
		return nil, err
	}
	return &JWTValidator{
		issuer:   issuer,
		audience: audience,
		JWKS:     jwks,
	}, nil
}

func (v JWTValidator) ValidateToken(request *http.Request) (map[string]interface{}, error) {
	tokenString := request.Header.Get("Authorization")
	splitToken := strings.Fields(tokenString)
	if strings.ToLower(splitToken[0]) != AuthTypeBearer {
		return nil, fmt.Errorf("unsupported authorization type, expected 'Bearer' %w", http.ErrNotSupported)
	}
	// Remove leading "Bearer "
	token := splitToken[len(splitToken)-1]
	result, err := jwt.Parse([]byte(token), jwt.WithKeySet(v.JWKS, jws.WithInferAlgorithmFromKey(true)))
	if err != nil {
		return nil, err
	}
	return result.PrivateClaims(), nil
}
