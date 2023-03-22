package auth

import (
	"github.com/Frontman-Labs/frontman/config"
	"net/http"
	"os"
	"testing"
)

func TestNewBasicAuthValidatorFromHardcodedCredentials(t *testing.T) {
	conf := &config.BasicAuthConfig{
		Username: "username",
		Password: "password",
	}

	validator, err := NewBasicAuthValidator(conf)
	if err != nil {
		t.Errorf("Failed to create basic validator: %s\n", err)
	}
	if validator.Username != "username" {
		t.Errorf("NewBasicAuthValidator failed to parse username from username config variable\n")
	}
	if validator.Password != "password" {
		t.Errorf("NewBasicAuthValidator failed to parse password from password config variable\n")
	}
}

func TestNewBasicAuthValidatorFromEnvVariables(t *testing.T) {
	conf := &config.BasicAuthConfig{
		UsernameEnv: "FRONTMAN_TEST_BACKEND_USERNAME",
		PasswordEnv: "FRONTMAN_TEST_BACKEND_PASSWORD",
	}

	os.Setenv("FRONTMAN_TEST_BACKEND_USERNAME", "username_from_env")
	os.Setenv("FRONTMAN_TEST_BACKEND_PASSWORD", "password_from_env")

	validator, err := NewBasicAuthValidator(conf)
	if err != nil {
		t.Errorf("Failed to create basic validator: %s\n", err)
	}
	if validator.Username != "username_from_env" {
		t.Errorf("NewBasicAuthValidator failed to parse username from username environment variable\n")
	}
	if validator.Password != "password_from_env" {
		t.Errorf("NewBasicAuthValidator failed to parse password from password environment variable\n")
	}
}

func TestBasicAuthValidCredentials(t *testing.T) {
	validator := &BasicAuthValidator{
		Username: "test",
		Password: "test",
	}

	req := &http.Request{
		Header: make(http.Header),
	}
	req.SetBasicAuth("test", "test")
	_, err := validator.ValidateToken(req)
	if err != nil {
		t.Errorf("Failed to validate correct basic auth: %s\n", err)
	}
}

func TestBasicAuthInvalidCredentials(t *testing.T) {
	validator := &BasicAuthValidator{
		Username: "test",
		Password: "test",
	}

	req := &http.Request{
		Header: make(http.Header),
	}
	req.SetBasicAuth("blah", "blah")
	_, err := validator.ValidateToken(req)
	if err == nil {
		t.Errorf("Failed to validate correctly identify invalid basic auth credentials\n")
	}

	if err.Error() != "Invalid credentials" {
		t.Errorf("Invalid error message returned when parsing invalid credentials: %s\n", err)
	}
}

func TestBasicAuthMissingCredentials(t *testing.T) {
	validator := &BasicAuthValidator{
		Username: "test",
		Password: "test",
	}

	req := &http.Request{
		Header: make(http.Header),
	}
	_, err := validator.ValidateToken(req)
	if err == nil {
		t.Errorf("Failed to validate correctly identify missing basic auth credentials\n")
	}

	if err.Error() != "Error parsing authentication token" {
		t.Errorf("Invalid error message returned when parsing invalid credentials: %s\n", err)
	}
}
