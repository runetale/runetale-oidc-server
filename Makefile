NO_CACHE =

.PHONY: build run css css-watch debug general migrate-db

build:
	docker build -f Dockerfile .
	docker build -f Dockerfile.db .
	docker-compose -f docker-compose.yml --profile db build $(NO_CACHE)
	docker-compose -f docker-compose.yml --profile general build $(NO_CACHE)

general:
	docker-compose --profile general up

migrate-db: build
	docker-compose --profile db up -d

dev: css migrate-db

run: css build migrate-db general

down:
	docker-compose --profile db down
	docker-compose --profile general down
	docker-compose --profile db down

css: npm-install
	tailwindcss -i css/input.css -o css/output.css --minify

css-watch: npm-install
	tailwindcss -i css/input.css -o css/output.css --watch

npm-install:
	npm install

nix-build:
	nix build .#runetale-oidc-server --no-link --print-out-paths --print-build-logs --extra-experimental-features flakes