package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type GoogleOAuthProvider struct {
	Config *oauth2.Config
}

type GoogleUserInfo struct {
	Sub       string `json:"sub"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	Locale    string `json:"locale"`
	ExpiresIn int64  `json:"expires_in"`
}

func NewGoogleOAuthProvider(clientID string, clientSecret string, redirectURL string, scopes []string) *GoogleOAuthProvider {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}

	return &GoogleOAuthProvider{Config: config}
}

func (p *GoogleOAuthProvider) GetAuthorizationURL(state string) string {
	return p.Config.AuthCodeURL(state)
}

func (p *GoogleOAuthProvider) ExchangeCodeForToken(code string, state string) (*oauth2.Token, error) {
	token, err := p.Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	if state != "" && token.Extra("state") != state {
		return nil, fmt.Errorf("invalid OAuth state")
	}

	return token, nil
}

func (p *GoogleOAuthProvider) GetUserInfo(token *oauth2.Token) (interface{}, error) {
	resp, err := http.Get(fmt.Sprintf("https://www.googleapis.com/oauth2/v2/userinfo?access_token=%s", token.AccessToken))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info, status: %s", resp.Status)
	}

	var userInfo GoogleUserInfo
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}
