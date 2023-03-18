package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"net/http"
	"testing"
	"time"
)

// TestGetServicesHandler tests the getServicesHandler function
func TestValidateJWTToken(t *testing.T) {
	validator := JWTValidator{
		JWKS: jwk.NewSet(),
	}
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate private key: %s\n", err)
	}
	// This is the key we will use to sign
	realKey, err := jwk.FromRaw(privKey)
	if err != nil {
		t.Errorf("failed to create JWK: %s\n", err)
	}

	realKey.Set(jwk.KeyIDKey, `mykey`)
	realKey.Set(jwk.AlgorithmKey, jwa.RS256)
	// add real key to the list of keys
	keyset := jwk.NewSet()
	keyset.AddKey(realKey)
	validator.JWKS, err = jwk.PublicSetOf(keyset)
	if err != nil {
		t.Errorf("Failed to generate public key set: %s", err)
	}
	authToken := jwt.New()
	authToken.Set(`email`, `ashketchum@gmail.com`)
	signed, err := jwt.Sign(authToken, jwt.WithKey(jwa.RS256, realKey))
	if err != nil {
		t.Errorf("failed to generate signed serialized: %s\n", err)
	}
	headers := make(http.Header)
	headers.Add("Authorization", string(signed))
	result, err := validator.ValidateToken(&http.Request{
		Header: headers,
	})
	if err != nil {
		t.Errorf("Failed to validate signed token: %s", err)
	}
	if result["email"].(string) != "ashketchum@gmail.com" {
		t.Errorf("Invalid email claim in parsed token: %s", result["email"].(string))
	}
}

func TestValidateJWTTokenInvalidSignature(t *testing.T) {
	validator := JWTValidator{
		JWKS: jwk.NewSet(),
	}
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate private key: %s\n", err)
	}
	// This is the key we will use to sign
	realKey, err := jwk.FromRaw(privKey)
	if err != nil {
		t.Errorf("failed to create JWK: %s\n", err)
	}

	realKey.Set(jwk.KeyIDKey, `mykey`)
	realKey.Set(jwk.AlgorithmKey, jwa.RS256)
	// add real key to the list of keys
	keyset := jwk.NewSet()
	if err != nil {
		t.Errorf("Failed to generate public key set: %s", err)
	}
	authToken := jwt.New()
	authToken.Set(`email`, `ashketchum@gmail.com`)
	badKey, err := jwk.FromRaw([]byte("test"))
	if err != nil {
		t.Errorf("Failed to generate bad key: %s", err)
	}
	keyset.AddKey(badKey)
	validator.JWKS, err = jwk.PublicSetOf(keyset)
	signed, err := jwt.Sign(authToken, jwt.WithKey(jwa.RS256, realKey))
	if err != nil {
		t.Errorf("failed to generate signed serialized: %s\n", err)
	}
	headers := make(http.Header)
	headers.Add("Authorization", string(signed))
	_, err = validator.ValidateToken(&http.Request{
		Header: headers,
	})
	if err == nil {
		t.Errorf("Failed to detect invalid key")
	}
}

func TestValidateJWTExpiredToken(t *testing.T) {
	validator := JWTValidator{
		JWKS: jwk.NewSet(),
	}
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate private key: %s\n", err)
	}
	// This is the key we will use to sign
	realKey, err := jwk.FromRaw(privKey)
	if err != nil {
		t.Errorf("failed to create JWK: %s\n", err)
	}

	realKey.Set(jwk.KeyIDKey, `mykey`)
	realKey.Set(jwk.AlgorithmKey, jwa.RS256)
	// add real key to the list of keys
	keyset := jwk.NewSet()
	if err != nil {
		t.Errorf("Failed to generate public key set: %s", err)
	}
	expired, err := time.Parse(time.RFC3339, "1995-06-07T15:04:05Z")
	if err != nil {
		t.Errorf("Failed to create expiration time: %s", err)
	}
	authToken := jwt.New()

	authToken.Set("exp", expired)
	authToken.Set(`email`, `ashketchum@gmail.com`)
	keyset.AddKey(realKey)
	validator.JWKS, err = jwk.PublicSetOf(keyset)
	signed, err := jwt.Sign(authToken, jwt.WithKey(jwa.RS256, realKey))
	if err != nil {
		t.Errorf("failed to generate signed serialized: %s\n", err)
	}
	headers := make(http.Header)
	headers.Add("Authorization", string(signed))
	_, err = validator.ValidateToken(&http.Request{
		Header: headers,
	})
	if err == nil {
		t.Errorf("Failed to detect invalid key: %s", err)
	}

	if err.Error() != "\"exp\" not satisfied" {
		t.Errorf("Failed to detect invalid expiration time")
	}
}
