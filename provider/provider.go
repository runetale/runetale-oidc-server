package provider

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/runetale/runetale-oidc-server/provider/oauth"
	"github.com/runetale/runetale-oidc-server/provider/oauth/github"
	"github.com/runetale/runetale-oidc-server/provider/oidc/google"
	"github.com/runetale/runetale-oidc-server/provider/oidc/microsoft"
	"golang.org/x/oauth2"
)

type Provider interface {
	GetRedirectURL(state, nonce string) string
	VerifyWithIDToken(rawIdToken string) (*oidc.IDToken, error)
	GetOAuth2Token(code string) (*oauth2.Token, error)
	// todo:(snt) refactor common user info
	GetUserInfo(token *oauth2.Token) (*oidc.UserInfo, error)
	GetOAuthUserInfo(token *oauth2.Token) (*oauth.UserInfo, error)
}

func NewProvider(o oauth.Options) (p Provider, err error) {
	ctx := context.Background()
	switch o.ProviderName {
	case google.Name:
		p, err = google.New(ctx, &o)
	case github.Name:
		p, err = github.New(ctx, &o)
	case microsoft.Name:
		p, err = microsoft.New(ctx, &o)
	default:
		return nil, fmt.Errorf("unknown provider: %s", o.ProviderName)
	}
	if err != nil {
		return nil, err
	}
	return p, err
}
