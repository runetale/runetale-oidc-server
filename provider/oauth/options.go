package oauth

import "net/url"

type Options struct {
	ProviderName string
	ProviderURL  string
	ClientID     string
	ClientSecret string
	RedirectURL  *url.URL
	TenantID     string
	Scopes       []string
}
