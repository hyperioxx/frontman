package frontman

import ("golang.org/x/oauth2")

type OAuthProvider interface {
    // GetAuthorizationURL returns the URL to redirect the user to for authentication
    GetAuthorizationURL(state string) string

    // ExchangeCodeForToken exchanges the authorization code received from the OAuth provider for an access token
    ExchangeCodeForToken(code string, state string) (*oauth2.Token, error)

    // GetUserInfo returns the user information associated with the access token
    GetUserInfo(token *oauth2.Token) (interface{}, error)
}