# runetale-oidc-server
runetale oidc/oauth provider.

# envriroment variables
This project uses environment variables to configure OAuth2 providers, database connections, and server behavior.
To get started, copy the `.env.example` file and create your own `.env` file:

```bash
cp .env.example .env
``` shell
# ========== Google OAuth2 ==========
GOOGLE_OAUTH2_CLIENT_ID=              # Google OAuth2 Client ID
GOOGLE_OAUTH2_CLIENT_SECRET=          # Google OAuth2 Client Secret
GOOGLE_OAUTH2_CALLBACK_URL=           # Redirect URI after Google authentication

# ========== GitHub OAuth2 ==========
GITHUB_OAUTH2_CLIENT_ID=              # GitHub OAuth2 Client ID
GITHUB_OAUTH2_CLIENT_SECRET=          # GitHub OAuth2 Client Secret
GITHUB_OAUTH2_CALLBACK_URL=           # Redirect URI after GitHub authentication

# ========== Microsoft OAuth2 ==========
MICROSOFT_OAUTH2_CLIENT_ID=           # Microsoft Azure OAuth2 Client ID
MICROSOFT_OAUTH2_CLIENT_SECRET=       # Microsoft OAuth2 Client Secret
MICROSOFT_OAUTH2_TENANT_ID=           # Azure Tenant ID
MICROSOFT_OAUTH2_CALLBACK_URL=        # Redirect URI after Microsoft authentication

# ========== PostgreSQL Configuration ==========
POSTGRESQL_URL=                       # PostgreSQL connection URL (e.g., for initialization)
POSTGRESQL_RUNE_OIDC_URL=             # PostgreSQL URL specifically for the `runetale-oidc-server` database

# ========== Application Database ==========
DB_URL=                               # URL used by the app to connect to the database
DB_NAME=                              # Name of the database (e.g., runetale-oidc-server)

# ========== gRPC Server ==========
RUNETALE_SERVER_URL=                  # Host and port for the Runetale Harmonize server (e.g., 127.0.0.1:50051)

# ========== Web Redirection After Login ==========
RUNETALE_WEB_REDIRECT_LOGIN_URL=     # Redirect URL for the web frontend after successful login (e.g., http://127.0.0.1:3000/authentication)

# ========== JWT Configuration ==========
JWT_SECRET=                           # Secret key used to sign JWTs
JWT_ISS=                              # JWT Issuer (e.g., "runetale")
JWT_AUD=                              # JWT Audience (e.g., "runetale")

# ========== TLS Configuration ==========
IS_TLS=                               # Whether to enable TLS (true/false)

# ========== Docker BuildKit ==========
DOCKER_BUILDKIT=1                     # Enables Docker BuildKit (improves build performance)
```

# setup develop enviroment
if you using nix flake, just like this.
``` sh
$ nix develop
```

# quick start
``` sh
$ make dev
$ go run main.go
```

or 
run with on docker container
``` sh
$ make run
```

# for nix
``` sh
$ make nix-build
```

# develop
## tailwind
this project uses tailwindcss embedded in go's templates engine.
we need to build tailwindcss.
``` sh
$ make css
```

## build
``` sh
$ make build
or
$ make build NO_CACHE=--no-cache
```

# Debugging DB
```sh
$ psql \
    --host=0.0.0.0 \
    --port=5433 \
    --username=postgres \
    --password \
    --dbname=runetale-oidc-server
```

# support providers
## oidc
- [x] google
- [ ] okta
- [x] ms

## oauth
- [x] github
- [ ] apple

## debug for peer to peer
If you are testing P2P locally using a VM or Docker, you will need to be creative in debugging due to the OIDC Provider.

1. launch runetale-oidc-server in both environments
Set runetale-oidc-server's runetale server environment variable to the IP of the host where you are running runetale-server.
example
`RUNETALE_SERVER_URL=172.16.165.130:8080`

2. change the environment variable of the runetale-admin-dashboard to
Please set the URL of envoy of runetale-server to the IP of the host where you are launching envoy.
example
`NEXT_PUBLIC_RUNETALE_SERVER_URL=http://172.16.165.130:8081`
