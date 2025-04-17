package github

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/runetale/runetale-oidc-server/crypto"
	"github.com/runetale/runetale-oidc-server/database"
	"github.com/runetale/runetale-oidc-server/entity"
	grpcclient "github.com/runetale/runetale-oidc-server/grpc_client"
	"github.com/runetale/runetale-oidc-server/provider"
	"github.com/runetale/runetale-oidc-server/provider/oauth"
	"github.com/runetale/runetale-oidc-server/provider/oauth/github"
	"github.com/runetale/runetale-oidc-server/repository"
	"google.golang.org/grpc"
)

var (
	maxTime = time.Unix(253370793661, 0) // year 9999
)

type GithubCallbackHandler struct {
	provider.Provider
	*database.Postgres

	tenant repository.TenantRepositoryImpl
	user   repository.UserRepositoryImpl

	grpcClient grpcclient.ServerClientImpl

	jwt crypto.JwtIssuer

	webRedirectLoginURL string
}

func NewCallbackHandler(clientid, clientsecret string,
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
		ProviderName: github.Name,
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

	return &GithubCallbackHandler{p, db, trepo, urepo, serverClient, crypto.NewJwtIssuer(jwtSecret, jwtAud, jwtIss), webRedirectLoginURL}, nil
}

func (g *GithubCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	oauth2Token, err := g.GetOAuth2Token(r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	oauth2Token.Expiry = maxTime

	userInfo, err := g.GetOAuthUserInfo(oauth2Token)
	if err != nil {
		http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	oauth2Token.AccessToken = "*REDACTED*"

	components := strings.Split(userInfo.Email, "@")
	_, domain := components[0], components[1]

	// db process
	// 1. if you were the first user that domain, create tenant.
	t, err := g.tenant.FindByDomain(domain)
	if errors.Is(err, database.ErrNoRows) {
		t = entity.NewTenant(domain)
		err = g.tenant.Create(t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tx.Commit()

		t, err = g.tenant.FindByDomain(domain)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// 2. create user with tenant id
	user := entity.NewUser(t.ID, userInfo.Profile, "", userInfo.Email, domain, userInfo.Subject, "", "", userInfo.Picture)
	err = g.user.Create(user)
	if err != nil && !errors.Is(err, database.ErrAlreadyExist) {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := g.jwt.CreateJwtTokenWithSub(userInfo.Subject)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get inviter
	_, _ = r.Cookie("inviter")

	// get invite_code
	inviteCode, _ := r.Cookie("invite_code")

	_, err = g.grpcClient.Login(user.Sub, t.TenantID, user.Domain, user.ProviderID, user.Email, user.Username, user.Picture, token, inviteCode.Value)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to Login "+err.Error(), http.StatusInternalServerError)
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
