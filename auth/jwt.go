package auth

import (
	"context"
	"log"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"net/http"
)

type JWTValidator struct {
	issuer   string
	audience string
	JWKS     jwk.Set
}

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
	// Remove leading "Bearer "
	token := splitToken[len(splitToken)-1]
	result, err := jwt.Parse([]byte(token), jwt.WithKeySet(v.JWKS, jws.WithInferAlgorithmFromKey(true)))
	if err != nil {
		return nil, err
	}
	return result.PrivateClaims(), nil
}
