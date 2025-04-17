package microsoft

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc"
	"github.com/runetale/runetale-oidc-server/provider/oauth"
	rune_oidc "github.com/runetale/runetale-oidc-server/provider/oidc"
)

const Name = "microsoft"

var (
	providerURL = "https://login.microsoftonline.com/%s/v2.0"
)

var defaultScopes = []string{oidc.ScopeOpenID, "profile", "email"}
var defaultAuthCodeOptions = map[string]string{"prompt": "select_account consent", "access_type": "offline"}

type Provider struct {
	*rune_oidc.Provider
}

// clientid, clientsecret and redirect url are set by the caller
func New(ctx context.Context, o *oauth.Options) (*Provider, error) {
	var p Provider

	o.ProviderURL = fmt.Sprintf(providerURL, o.TenantID)
	o.Scopes = defaultScopes

	oidc, err := rune_oidc.New(ctx, o)
	p.Provider = oidc

	return &p, err
}
