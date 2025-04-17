package oidc

import (
	"context"
	"errors"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/runetale/runetale-oidc-server/provider/oauth"
	"golang.org/x/oauth2"
)

type Provider struct {
	config   oauth2.Config
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier

	ctx context.Context
}

func New(ctx context.Context, o *oauth.Options) (*Provider, error) {
	p := new(Provider)

	np, err := oidc.NewProvider(ctx, o.ProviderURL)
	if err != nil {
		return nil, err
	}
	p.provider = np

	config := oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		Endpoint:     np.Endpoint(),
		RedirectURL:  o.RedirectURL.String(),
		Scopes:       o.Scopes,
	}
	p.config = config

	oidcConfig := &oidc.Config{
		ClientID: o.ClientID,
	}

	verifier := np.Verifier(oidcConfig)
	p.verifier = verifier

	p.ctx = ctx

	return p, nil
}

func (p *Provider) GetRedirectURL(state, nonce string) string {
	return p.config.AuthCodeURL(state, oidc.Nonce(nonce))
}

func (p *Provider) VerifyWithIDToken(rawIdToken string) (*oidc.IDToken, error) {
	return p.verifier.Verify(p.ctx, rawIdToken)
}

func (p *Provider) GetOAuth2Token(code string) (*oauth2.Token, error) {
	return p.config.Exchange(p.ctx, code)
}

func (p *Provider) GetUserInfo(token *oauth2.Token) (*oidc.UserInfo, error) {
	return p.provider.UserInfo(p.ctx, oauth2.StaticTokenSource(token))
}

func (p *Provider) GetOAuthUserInfo(token *oauth2.Token) (*oauth.UserInfo, error) {
	return nil, errors.New("not implemented")
}
