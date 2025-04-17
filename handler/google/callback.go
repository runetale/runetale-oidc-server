package google

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/runetale/runetale-oidc-server/crypto"
	"github.com/runetale/runetale-oidc-server/database"
	"github.com/runetale/runetale-oidc-server/entity"
	grpcclient "github.com/runetale/runetale-oidc-server/grpc_client"
	"github.com/runetale/runetale-oidc-server/provider"
	"github.com/runetale/runetale-oidc-server/provider/oauth"
	"github.com/runetale/runetale-oidc-server/provider/oidc/google"
	"github.com/runetale/runetale-oidc-server/repository"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

type GoogleCallbackHandler struct {
	provider.Provider
	*database.Postgres

	tenant repository.TenantRepositoryImpl
	user   repository.UserRepositoryImpl

	grpcClient grpcclient.ServerClientImpl

	jwt crypto.JwtIssuer

	webRedirectLoginURL string
}

func NewCallbackHandler(
	clientid, clientsecret string,
	jwtSecret, jwtAud, jwtIss string,
	webRedirectLoginURL string,
	callbackURL string,
	db *database.Postgres, conn *grpc.ClientConn,
) (Handler, error) {
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

	trepo := repository.NewTenantRepository(db)
	urepo := repository.NewUserRepository(db)

	serverClient := grpcclient.NewServerClient(conn)

	return &GoogleCallbackHandler{p, db, trepo, urepo, serverClient, crypto.NewJwtIssuer(jwtSecret, jwtAud, jwtIss), webRedirectLoginURL}, nil
}

func (g *GoogleCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tx, err := g.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("can not begin to transaction: %s", err.Error()), http.StatusBadRequest)
	}

	state, err := r.Cookie("state")
	if err != nil {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	// 1. get oauth2Token from query of code
	oauth2Token, err := g.GetOAuth2Token(r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. get rawIDToken from id_token of oauth2Token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}

	// 3. get idToken, after go_oidc verifier verified to rawIDToken
	idToken, err := g.VerifyWithIDToken(rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. check nonce of idToken
	nonce, err := r.Cookie("nonce")
	if err != nil {
		http.Error(w, "nonce not found", http.StatusBadRequest)
		return
	}
	if idToken.Nonce != nonce.Value {
		http.Error(w, "nonce did not match", http.StatusBadRequest)
		return
	}

	// get userinfo
	userInfo, err := g.GetUserInfo(oauth2Token)
	if err != nil {
		http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	oauth2Token.AccessToken = "*REDACTED*"

	// get idToken claims
	type idTokenClaims struct {
		Iss string `json:"iss"`
		Azp string `json:"azp"`
		Aud string `json:"aud"`
		Sub string `json:"sub"`
		// domain
		Hd            string `json:"hd"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		AtHash        string `json:"at_hash"`
		Nonce         string `json:"nonce"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Locale        string `json:"en"`
		Iat           int64  `json:"iat"`
		Exp           int64  `json:"exp"`
	}
	itc := idTokenClaims{}
	resp := struct {
		OAuth2Token   *oauth2.Token
		IDTokenClaims *idTokenClaims
	}{oauth2Token, &itc}

	if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// db process
	// 1. if you were the first user that domain, create tenant.
	t, err := g.tenant.FindByDomain(itc.Hd)
	if errors.Is(err, database.ErrNoRows) {
		t = entity.NewTenant(itc.Hd)
		err = g.tenant.Create(t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tx.Commit()

		t, err = g.tenant.FindByDomain(itc.Hd)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// 2. create user with tenant id
	user := entity.NewUser(t.ID, itc.Name, itc.Aud, itc.Email, itc.Hd, userInfo.Subject, itc.Aud, itc.Azp, itc.Picture)
	err = g.user.Create(user)
	if err != nil && !errors.Is(err, database.ErrAlreadyExist) {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := g.jwt.CreateJwtTokenWithSub(itc.Sub)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get inviter
	_, _ = r.Cookie("inviter")

	// get invite_code
	inviteCode, _ := r.Cookie("invite_code")

	// 3. login to server
	_, err = g.grpcClient.Login(user.Sub, t.TenantID, user.Domain, user.ProviderID, user.Email, user.Username, user.Picture, token, inviteCode.Value)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to Login,"+err.Error(), http.StatusInternalServerError)
		return
	}

	tx.Commit()

	_, _, err = g.jwt.GetCustomClamis(token, g.jwt.GetJwtSecret())
	if err != nil {
		http.Error(w, "Failed to Claims "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to true if using HTTPS
	})

	// 5. redirect to web admin URL with jwt token
	// the redirected web page uses a cookie sub and issues a jwt token on the web application side.
	http.Redirect(w, r, g.webRedirectLoginURL+"?token="+token, http.StatusFound)

}
