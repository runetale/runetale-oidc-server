package google

import (
	"context"

	"github.com/coreos/go-oidc"
	"github.com/runetale/runetale-oidc-server/provider/oauth"
	rune_oidc "github.com/runetale/runetale-oidc-server/provider/oidc"
)

const Name = "google"

var (
	providerURL = "https://accounts.google.com"
)

var defaultScopes = []string{oidc.ScopeOpenID, "profile", "email"}
var defaultAuthCodeOptions = map[string]string{"prompt": "select_account consent", "access_type": "offline"}

type Provider struct {
	*rune_oidc.Provider
}

// clientid, clientsecret and redirect url are set by the caller
func New(ctx context.Context, o *oauth.Options) (*Provider, error) {
	var p Provider

	o.ProviderURL = providerURL
	o.Scopes = defaultScopes

	oidc, err := rune_oidc.New(ctx, o)
	p.Provider = oidc

	return &p, err
}
