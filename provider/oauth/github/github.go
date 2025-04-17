package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	httpclient "github.com/runetale/runetale-oidc-server/http_request"
	"github.com/runetale/runetale-oidc-server/provider/oauth"
	"golang.org/x/oauth2"
	oauth2github "golang.org/x/oauth2/github"
	"gopkg.in/square/go-jose.v2/jwt"
)

const Name = "github"

const (
	refreshExpiry = time.Minute * 60
)

var (
	defaultProviderURL = "https://github.com"
	userEndpoint       = "https://api.github.com/user"
	emailEndpoint      = "https://api.github.com/user/emails"
)

var defaultScopes = []string{"user:email", "read:org"}

type Provider struct {
	Oauth *oauth2.Config

	userEndPoint  string
	emailEndPoint string

	ctx context.Context
}

// clientid, clientsecret and redirect url are set by the caller
func New(ctx context.Context, o *oauth.Options) (*Provider, error) {
	var p Provider

	p.ctx = ctx
	p.Oauth = &oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		Scopes:       defaultScopes,
		RedirectURL:  o.RedirectURL.String(),
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauth2github.Endpoint.AuthURL,
			TokenURL: oauth2github.Endpoint.TokenURL,
		},
	}

	return &p, nil
}

func (p *Provider) GetRedirectURL(state, nonce string) string {
	return p.Oauth.AuthCodeURL(state, oidc.Nonce(nonce))
}

func (p *Provider) VerifyWithIDToken(rawIdToken string) (*oidc.IDToken, error) {
	return nil, errors.New(fmt.Sprintf("%s provider does not support the function", Name))
}

func (p *Provider) GetOAuth2Token(code string) (*oauth2.Token, error) {
	return p.Oauth.Exchange(p.ctx, code)
}

type SessionClaims struct {
	Claims
	RawIDToken string
}

// Claims are JWT claims.
type Claims map[string]interface{}

type GithubUserResponse struct {
	Subject string `json:"sub"`
	Name    string `json:"name,omitempty"`
	User    string `json:"user"`
	Picture string `json:"picture,omitempty"`
	// needs to be set manually
	Expiry    *jwt.NumericDate `json:"exp,omitempty"`
	NotBefore *jwt.NumericDate `json:"nbf,omitempty"`
	IssuedAt  *jwt.NumericDate `json:"iat,omitempty"`
}

func (p *Provider) GetUserInfo(token *oauth2.Token) (*oidc.UserInfo, error) {
	var response struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url,omitempty"`
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("token %s", token.AccessToken),
		"Accept":        "application/vnd.github.v3+json",
	}

	version := fmt.Sprintf("%s/%s (+%s; %s; %s)", "", "", "", "", runtime.Version())
	err := httpclient.Do(p.ctx, http.MethodGet, p.userEndPoint, version, headers, nil, &response)
	if err != nil {
		return nil, err
	}

	var out GithubUserResponse

	out.Expiry = jwt.NewNumericDate(time.Now().Add(refreshExpiry))
	out.NotBefore = jwt.NewNumericDate(time.Now())
	out.IssuedAt = jwt.NewNumericDate(time.Now())

	out.User = response.Login
	out.Subject = response.Login
	out.Name = response.Name
	out.Picture = response.AvatarURL
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	var claims SessionClaims
	err = json.Unmarshal(b, &claims)
	if err != nil {
		return nil, err
	}

	email, err := p.getUserEmail(token)

	return &oidc.UserInfo{
		Subject:       out.Subject,
		Profile:       out.Picture,
		Email:         email.Email,
		EmailVerified: email.Verified,
	}, nil
}

type GithubEmailResponse struct {
	Email    string `json:"email"`
	Verified bool   `json:"email_verified"`
}

func (p *Provider) getUserEmail(t *oauth2.Token) (*GithubEmailResponse, error) {
	var response []struct {
		Email      string `json:"email"`
		Verified   bool   `json:"verified"`
		Primary    bool   `json:"primary"`
		Visibility string `json:"visibility"`
	}

	headers := map[string]string{"Authorization": fmt.Sprintf("token %s", t.AccessToken)}

	version := fmt.Sprintf("%s/%s (+%s; %s; %s)", "", "", "", "", runtime.Version())
	err := httpclient.Do(p.ctx, http.MethodGet, p.emailEndPoint, version, headers, nil, &response)
	if err != nil {
		return nil, err
	}

	var out GithubEmailResponse
	for _, email := range response {
		if email.Primary && email.Verified {
			out.Email = email.Email
			out.Verified = true
			break
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	var claims SessionClaims
	err = json.Unmarshal(b, &claims)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (p *Provider) GetOAuthUserInfo(token *oauth2.Token) (*oauth.UserInfo, error) {
	var response struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url,omitempty"`
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("token %s", token.AccessToken),
		"Accept":        "application/vnd.github.v3+json",
	}

	version := fmt.Sprintf("%s/%s (+%s; %s; %s)", "", "", "", "", runtime.Version())
	err := httpclient.Do(p.ctx, http.MethodGet, p.userEndPoint, version, headers, nil, &response)
	if err != nil {
		return nil, err
	}

	var out GithubUserResponse

	out.Expiry = jwt.NewNumericDate(time.Now().Add(refreshExpiry))
	out.NotBefore = jwt.NewNumericDate(time.Now())
	out.IssuedAt = jwt.NewNumericDate(time.Now())

	out.User = response.Login
	out.Subject = response.Login
	out.Name = response.Name
	out.Picture = response.AvatarURL
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	var claims SessionClaims
	err = json.Unmarshal(b, &claims)
	if err != nil {
		return nil, err
	}

	email, err := p.getUserEmail(token)

	return &oauth.UserInfo{
		Name:          out.Name,
		User:          out.User,
		Picture:       out.Picture,
		Subject:       out.Subject,
		Profile:       out.Picture,
		Email:         email.Email,
		EmailVerified: email.Verified,
	}, nil
}
