.PHONY: build build-backend build-frontend build-proxy push-backend push-frontend push-proxy push-all up down logs

SERVICES := backend frontend proxy
GHCR_PREFIX := ghcr.io/ice-rider/telegleb

# Build all images locally
build: build-backend build-frontend build-proxy

build-backend:
	docker build -t $(GHCR_PREFIX)/backend:latest ./server

build-frontend:
	docker build -t $(GHCR_PREFIX)/frontend:latest ./client

build-proxy:
	docker build -t $(GHCR_PREFIX)/proxy:latest ./proxy

# Push all images to GHCR
push-all: push-backend push-frontend push-proxy

push-backend:
	docker push $(GHCR_PREFIX)/backend:latest

push-frontend:
	docker push $(GHCR_PREFIX)/frontend:latest

push-proxy:
	docker push $(GHCR_PREFIX)/proxy:latest

# Docker Compose (dev)
up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

# Docker Compose (prod)
prod-up:
	docker compose -f docker-compose.prod.yaml up -d

prod-down:
	docker compose -f docker-compose.prod.yaml down

# Git helpers
commit:
	git add -A && git commit -m "$(msg)"

push:
	git push origin main
