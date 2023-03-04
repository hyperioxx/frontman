package frontman

import (
	"context"
	"encoding/json"

	"golang.org/x/oauth2"
)



type KeycloakProvider struct {
    clientID     string
    clientSecret string
    redirectURI  string
    authURL      string
    tokenURL     string
    userinfoURL  string
}

func NewKeycloakProvider(clientID, clientSecret, redirectURI, authURL, tokenURL, userinfoURL string) *KeycloakProvider {
    return &KeycloakProvider{
        clientID:     clientID,
        clientSecret: clientSecret,
        redirectURI:  redirectURI,
        authURL:      authURL,
        tokenURL:     tokenURL,
        userinfoURL:  userinfoURL,
    }
}

func (kp *KeycloakProvider) GetAuthorizationURL(state string) string {
    conf := &oauth2.Config{
        ClientID:     kp.clientID,
        ClientSecret: kp.clientSecret,
        RedirectURL:  kp.redirectURI,
        Endpoint: oauth2.Endpoint{
            AuthURL:  kp.authURL,
            TokenURL: kp.tokenURL,
        },
        Scopes: []string{"openid", "profile", "email"},
    }

    return conf.AuthCodeURL(state)
}

func (kp *KeycloakProvider) ExchangeCodeForToken(code string, state string) (*oauth2.Token, error) {
    conf := &oauth2.Config{
        ClientID:     kp.clientID,
        ClientSecret: kp.clientSecret,
        RedirectURL:  kp.redirectURI,
        Endpoint: oauth2.Endpoint{
            AuthURL:  kp.authURL,
            TokenURL: kp.tokenURL,
        },
        Scopes: []string{"openid", "profile", "email"},
    }

    token, err := conf.Exchange(oauth2.NoContext, code)
    if err != nil {
        return nil, err
    }

    return token, nil
}

func (kp *KeycloakProvider) GetUserInfo(token *oauth2.Token) (interface{}, error) {
    httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
    resp, err := httpClient.Get(kp.userinfoURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var userInfo map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&userInfo)
    if err != nil {
        return nil, err
    }

    return userInfo, nil
}
