package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/runetale/runetale-oidc-server/database"
	grpcclient "github.com/runetale/runetale-oidc-server/grpc_client"
	"github.com/runetale/runetale-oidc-server/handler/github"
	"github.com/runetale/runetale-oidc-server/handler/google"
	"github.com/runetale/runetale-oidc-server/handler/invite"
	"github.com/runetale/runetale-oidc-server/handler/microsoft"
	"google.golang.org/grpc"
)

var (
	//go:embed all:templates/*
	templateFS embed.FS

	//go:embed all:assets/*
	assetsFS embed.FS

	//go:embed all:fonts/*
	fontsFS embed.FS

	//go:embed css/output.css
	css embed.FS

	//parsed templates
	html *template.Template
)

var (
	googleClientID        = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID")
	googleClientSecret    = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
	googleCallbackURL     = os.Getenv("GOOGLE_OAUTH2_CALLBACK_URL")
	githubClientID        = os.Getenv("GITHUB_OAUTH2_CLIENT_ID")
	githubClientSecret    = os.Getenv("GITHUB_OAUTH2_CLIENT_SECRET")
	githubCallbackURL     = os.Getenv("GITHUB_OAUTH2_CALLBACK_URL")
	microsoftClientID     = os.Getenv("MICROSOFT_OAUTH2_CLIENT_ID")
	microsoftClientSecret = os.Getenv("MICROSOFT_OAUTH2_CLIENT_SECRET")
	microsoftTenantID     = os.Getenv("MICROSOFT_OAUTH2_TENANT_ID")
	microsoftCallbackURL  = os.Getenv("MICROSOFT_OAUTH2_CALLBACK_URL")
	dburl                 = os.Getenv("DB_URL")
	isTLS                 = os.Getenv("IS_TLS")
	runetaleServerURL     = os.Getenv("RUNETALE_SERVER_URL")

	jwtSecret = os.Getenv("JWT_SECRET")
	jwtIss    = os.Getenv("JWT_ISS")
	jwtAud    = os.Getenv("JWT_AUD")

	webRedirectLoginURL = os.Getenv("RUNETALE_WEB_REDIRECT_LOGIN_URL")
)

func templateParseFSRecursive(
	templates fs.FS,
	ext string,
	nonRootTemplateNames bool,
	funcMap template.FuncMap) (*template.Template, error) {

	root := template.New("")
	err := fs.WalkDir(templates, "templates", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ext) {
			if err != nil {
				return err
			}
			b, err := fs.ReadFile(templates, path)
			if err != nil {
				return err
			}
			name := ""
			if nonRootTemplateNames {
				//name the template based on the file path (excluding the root)
				parts := strings.Split(path, string(os.PathSeparator))
				name = strings.Join(parts[1:], string(os.PathSeparator))
			}
			t := root.New(name).Funcs(funcMap)
			_, err = t.Parse(string(b))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return root, err
}

type Response struct {
	Status      int
	ContentType string
	Content     io.Reader
	Headers     map[string]string
}

func responsehtml(status int, t *template.Template, template string, data interface{}, headers map[string]string) *Response {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, template, data); err != nil {
		log.Println(err)
		return nil
	}
	return &Response{
		Status:      status,
		ContentType: "text/html",
		Content:     &buf,
		Headers:     headers,
	}
}

func index(r *http.Request) *Response {
	return responsehtml(http.StatusOK, html, "index.html", nil, nil)
}

type Action func(r *http.Request) *Response

func (a Action) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	response := a(r)
	response.Write(rw)
}

func (response *Response) Write(rw http.ResponseWriter) {
	if response != nil {
		if response.ContentType != "" {
			rw.Header().Set("Content-Type", response.ContentType)
		}
		for k, v := range response.Headers {
			rw.Header().Set(k, v)
		}
		rw.WriteHeader(response.Status)
		_, err := io.Copy(rw, response.Content)

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		rw.WriteHeader(http.StatusOK)
	}
}

func newPostgres() (*database.Postgres, error) {
	db, err := database.NewPostgres(dburl)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, err
}

func main() {
	var err error
	isTLS, err := strconv.ParseBool(isTLS)
	if err != nil {
		isTLS = false
	}

	html, err = templateParseFSRecursive(templateFS, ".html", true, nil)
	if err != nil {
		return
	}

	db, err := newPostgres()
	if err != nil {
		fmt.Println(err)
		return
	}

	op := grpcclient.NewGrpcDialOption(isTLS)
	clientCtx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		clientCtx,
		runetaleServerURL,
		op,
		grpc.WithBlock(),
	)
	if err != nil {
		panic(err)
	}

	// have to set middleware handler(more security)
	router := http.NewServeMux()
	// web
	router.Handle("/css/output.css", http.FileServer(http.FS(css)))
	router.Handle("/assets/", http.FileServer(http.FS(assetsFS)))
	router.Handle("/fonts/", http.FileServer(http.FS(fontsFS)))
	router.Handle("/", Action(index))

	// server
	// invit
	inviteHandler := invite.NewInviteHandler(conn, html)
	router.Handle("/invite", inviteHandler)

	// google
	googleHandler, err := google.NewGoogleHandler(googleClientID, googleClientSecret, googleCallbackURL, db)
	if err != nil {
		panic(err)
	}
	router.Handle("/google", googleHandler)
	googleCallbackHandler, err := google.NewCallbackHandler(googleClientID, googleClientSecret, jwtSecret, jwtAud, jwtIss, webRedirectLoginURL, googleCallbackURL, db, conn)
	if err != nil {
		panic(err)
	}
	router.Handle("/auth/google/callback", googleCallbackHandler)

	// github
	githubHandler, err := github.NewGithubHandler(githubClientID, githubClientSecret, githubCallbackURL, db)
	if err != nil {
		panic(err)
	}
	router.Handle("/github", githubHandler)
	githubCallbackHandler, err := github.NewCallbackHandler(githubClientID, githubClientSecret, jwtSecret, jwtAud, jwtIss, webRedirectLoginURL, githubCallbackURL, db, conn)
	if err != nil {
		panic(err)
	}
	router.Handle("/auth/github/callback", githubCallbackHandler)

	// microsoft
	microsoftHandler, err := microsoft.NewMicrosoftHandler(microsoftClientID, microsoftClientSecret, microsoftTenantID, microsoftCallbackURL, db)
	if err != nil {
		panic(err)
	}
	router.Handle("/microsoft", microsoftHandler)
	microsoftCallbackHandler, err := microsoft.NewCallbackHandler(microsoftClientID, microsoftClientSecret, microsoftTenantID, jwtSecret, jwtAud, jwtIss, webRedirectLoginURL, microsoftCallbackURL, db, conn)
	if err != nil {
		panic(err)
	}
	router.Handle("/auth/microsoft/callback", microsoftCallbackHandler)

	log.Printf("listening on :5556")
	log.Fatal(http.ListenAndServe("0.0.0.0:5556", router))
}
