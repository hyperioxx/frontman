package auth

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Frontman-Labs/frontman/config"
	"gopkg.in/yaml.v3"
)

type BasicAuthValidator struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func getCredentialsFromConfig(conf *config.BasicAuthConfig) (string, string) {
	var username, password string
	if conf.Username != "" {
		username = conf.Username
	} else {
		username = os.Getenv(conf.UsernameEnv)
	}

	if conf.Password != "" {
		password = conf.Password
	} else {
		password = os.Getenv(conf.PasswordEnv)
	}

	return username, password
}

func NewBasicAuthValidator(conf *config.BasicAuthConfig) (*BasicAuthValidator, error) {
	if conf.CredentialsFile != "" {
		// Read credentials file to build validator
		yamlData, err := ioutil.ReadFile(conf.CredentialsFile)
		if err != nil {
			log.Printf("Failed to read credentials file: %s", err)
			return nil, err
		}
		validator := &BasicAuthValidator{}
		err = yaml.Unmarshal(yamlData, validator)
		if err != nil {
			log.Printf("Failed to unmarshal credentials data: %s", err)
			return nil, err
		}
		return validator, nil
	}
	username, password := getCredentialsFromConfig(conf)
	return &BasicAuthValidator{
		Username: username,
		Password: password,
	}, nil
}

func (v BasicAuthValidator) ValidateToken(request *http.Request) (map[string]interface{}, error) {
	username, password, ok := request.BasicAuth()
	if !ok {
		return nil, errors.New("Error parsing authentication token")
	}

	if username != v.Username || password != v.Password {
		return nil, errors.New("Invalid credentials")
	}

	return nil, nil
}
