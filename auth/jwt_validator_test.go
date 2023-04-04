package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Frontman-Labs/frontman/config"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

func buildKey(t *testing.T, alg jwa.SignatureAlgorithm, KeySize int) jwk.Key {
	var newKey jwk.Key
	switch alg {
	case jwa.RS256:

		privKey, err := rsa.GenerateKey(rand.Reader, KeySize)
		require.NoError(t, err)
		// This is the key we will use to sign
		jwKey, err := jwk.FromRaw(privKey)
		require.NoError(t, err)
		jwKey.Set(jwk.KeyIDKey, `mykey`)
		jwKey.Set(jwk.AlgorithmKey, alg)
		newKey = jwKey
	}
	return newKey
}

func buildDefaultToken(t *testing.T, claims map[string]interface{}) jwt.Token {
	token := jwt.New()

	for key, value := range claims {
		err := token.Set(key, value)
		require.NoError(t, err)
	}
	fmt.Println(token.Expiration())
	return token
}

// TestGetServicesHandler tests the getServicesHandler function
func TestValidateJWTToken(t *testing.T) {
	defaultClaims := map[string]interface{}{
		"aud":   "frontman.com",
		"email": "ashketchum@gmail.com",
		"exp":   time.Now().Add(time.Hour),
		"iat":   time.Now(),
		"iss":   "frontman.com",
		"sub":   "ashketchum@gmail.com",
	}
	testCases := []struct {
		description     string
		buildAuthHeader func(t *testing.T, signingKey jwk.Key, alg jwa.SignatureAlgorithm) *http.Request
		buildValidator  func(t *testing.T, validatingKey jwk.Key) *JWTValidator
		validateResult  func(t *testing.T, result map[string]interface{}, err error)
	}{
		{
			description: "OK",
			buildAuthHeader: func(t *testing.T, signingKey jwk.Key, alg jwa.SignatureAlgorithm) *http.Request {
				authToken := buildDefaultToken(t, defaultClaims)
				signedAuthToken, err := jwt.Sign(authToken, jwt.WithKey(alg, signingKey))
				require.NoError(t, err)

				headers := make(http.Header)
				headers.Add("Authorization", fmt.Sprintf("Bearer %s", string(signedAuthToken)))
				return &http.Request{
					Header: headers,
				}
			},
			validateResult: func(t *testing.T, result map[string]interface{}, err error) {
				require.NoError(t, err)
				// test private claims
				require.Equal(t, result["email"].(string), defaultClaims["email"])
			},
		},
		{
			description: "empty auth header",
			buildAuthHeader: func(t *testing.T, signingKey jwk.Key, alg jwa.SignatureAlgorithm) *http.Request {
				headers := make(http.Header)
				headers.Add("Authorization", "")
				return &http.Request{
					Header: headers,
				}
			},
			validateResult: func(t *testing.T, result map[string]interface{}, err error) {
				require.ErrorIs(t, err, ErrMissingAuthHeader)
			},
		},
		{
			description: "absent auth header",
			buildAuthHeader: func(t *testing.T, signingKey jwk.Key, alg jwa.SignatureAlgorithm) *http.Request {
				return &http.Request{}
			},
			validateResult: func(t *testing.T, result map[string]interface{}, err error) {
				require.ErrorIs(t, err, ErrMissingAuthHeader)
			},
		},
		{
			description: "bad format type for auth header",
			buildAuthHeader: func(t *testing.T, signingKey jwk.Key, alg jwa.SignatureAlgorithm) *http.Request {
				headers := make(http.Header)
				headers.Add("Authorization", "unknownAuthType mytoken")
				return &http.Request{
					Header: headers,
				}
			},
			validateResult: func(t *testing.T, result map[string]interface{}, err error) {
				require.ErrorIs(t, err, http.ErrNotSupported)
			},
		},
		{
			description: "invalid format for auth header",
			buildAuthHeader: func(t *testing.T, signingKey jwk.Key, alg jwa.SignatureAlgorithm) *http.Request {
				headers := make(http.Header)
				headers.Add("Authorization", "tokenwithoutschema")
				return &http.Request{
					Header: headers,
				}
			},
			validateResult: func(t *testing.T, result map[string]interface{}, err error) {
				require.ErrorIs(t, err, ErrBadFormatAuthHeader)
			},
		},
		{
			description: "invalid signature",
			buildAuthHeader: func(t *testing.T, signingKey jwk.Key, alg jwa.SignatureAlgorithm) *http.Request {
				authToken := buildDefaultToken(t, defaultClaims)
				badKey := buildKey(t, alg, 2048)
				signedAuthToken, err := jwt.Sign(authToken, jwt.WithKey(alg, badKey))
				require.NoError(t, err)

				headers := make(http.Header)
				headers.Add("Authorization", fmt.Sprintf("Bearer %s", string(signedAuthToken)))
				return &http.Request{
					Header: headers,
				}
			},
			validateResult: func(t *testing.T, result map[string]interface{}, err error) {
				// Expect the forged token to be rejected
				require.Error(t, err)
			},
		},
		{
			description: "expired token",
			buildAuthHeader: func(t *testing.T, signingKey jwk.Key, alg jwa.SignatureAlgorithm) *http.Request {
				
				customClaims := make(map[string]interface{})
				maps.Copy(defaultClaims, customClaims)
				customClaims["exp"] = time.Now().Add(-10 * time.Minute)
				authToken := buildDefaultToken(t, customClaims)
				signedAuthToken, err := jwt.Sign(authToken, jwt.WithKey(alg, signingKey))
				require.NoError(t, err)

				headers := make(http.Header)
				headers.Add("Authorization", fmt.Sprintf("Bearer %s", string(signedAuthToken)))
				return &http.Request{
					Header: headers,
				}
			},
			validateResult: func(t *testing.T, result map[string]interface{}, err error) {
				// Expect the expired token to be rejected
				require.Error(t, err)
			},
		},
		{
			description: "OK + http-fetched JWK",
			buildAuthHeader: func(t *testing.T, signingKey jwk.Key, alg jwa.SignatureAlgorithm) *http.Request {
				authToken := buildDefaultToken(t, defaultClaims)
				signedAuthToken, err := jwt.Sign(authToken, jwt.WithKey(alg, signingKey))
				require.NoError(t, err)

				headers := make(http.Header)
				headers.Add("Authorization", fmt.Sprintf("Bearer %s", string(signedAuthToken)))
				return &http.Request{
					Header: headers,
				}
			},
			buildValidator: func(t *testing.T, validatingKey jwk.Key) *JWTValidator {
				set := jwk.NewSet()
				set.AddKey(validatingKey)
				validatingKeySet, err := jwk.PublicSetOf(set)
				returnedKey, err := json.Marshal(validatingKeySet)
				require.NoError(t, err)

				// build mock server to serve jwks
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write(returnedKey)
				}))
				v, err := NewJWTValidator(&config.JWTConfig{KeysUrl: srv.URL})
				require.NoError(t, err)
				return v
			},
			validateResult: func(t *testing.T, result map[string]interface{}, err error) {
				require.NoError(t, err)
			},
		},
		{
			description: "bad jwks url",
			buildAuthHeader: func(t *testing.T, signingKey jwk.Key, alg jwa.SignatureAlgorithm) *http.Request {
				return &http.Request{}
			},
			buildValidator: func(t *testing.T, validatingKey jwk.Key) *JWTValidator {
				_, err := NewJWTValidator(&config.JWTConfig{KeysUrl: "http://localhost/invalid/jwk/endpoint"})
				require.Error(t, err)
				return &JWTValidator{}
			},
			validateResult: func(t *testing.T, result map[string]interface{}, err error) {
				// expect to fail because no jwk was loaded
				require.Error(t, err)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.description, func(t *testing.T) {
			alg := jwa.RS256
			key := buildKey(t, alg, 2048)
			keyset := jwk.NewSet()
			keyset.AddKey(key)
			JWKS, err := jwk.PublicSetOf(keyset)
			require.NoError(t, err)
			if tc.buildValidator == nil {
				// Default validator builder with manually generated JWKS
				tc.buildValidator = func(t *testing.T, validatingKey jwk.Key) *JWTValidator {
					v, err := NewJWTValidator(&config.JWTConfig{}, func(j *JWTValidator) { j.JWKS = JWKS })
					require.NoError(t, err)
					return v
				}
			}
			validator := tc.buildValidator(t, key)
			result, err := validator.ValidateToken(tc.buildAuthHeader(t, key, alg))
			tc.validateResult(t, result, err)
		})
	}
}
