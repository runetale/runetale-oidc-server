package google

import (
	"net/http"
	"net/url"

	"github.com/runetale/runetale-oidc-server/database"
	"github.com/runetale/runetale-oidc-server/handler"
	"github.com/runetale/runetale-oidc-server/provider"
	"github.com/runetale/runetale-oidc-server/provider/oauth"
	"github.com/runetale/runetale-oidc-server/provider/oidc/google"
)

type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type GoogleHandler struct {
	provider.Provider
}

func NewGoogleHandler(clientid, clientsecret, callbackURL string, db *database.Postgres) (Handler, error) {
	redirectURL, err := url.Parse(callbackURL)
	if err != nil {
		return nil, err
	}
	options := oauth.Options{
		ProviderName: google.Name,
		ClientID:     clientid,
		ClientSecret: clientsecret,
		RedirectURL:  redirectURL,
	}
	p, err := provider.NewProvider(options)
	if err != nil {
		return nil, err
	}

	return &GoogleHandler{p}, nil
}

func (g *GoogleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inviter := r.URL.Query().Get("inviter")
	inviteCode := r.URL.Query().Get("invite_code")

	state, err := handler.RandString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	nonce, err := handler.RandString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	handler.SetCallbackCookie(w, r, "state", state)
	handler.SetCallbackCookie(w, r, "nonce", nonce)
	handler.SetCallbackCookie(w, r, "inviter", inviter)
	handler.SetCallbackCookie(w, r, "invite_code", inviteCode)

	http.Redirect(w, r, g.GetRedirectURL(state, nonce), http.StatusFound)
}
