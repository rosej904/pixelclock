BINARY       := ulanzi-clock
CMD          := ./cmd/clock
IMAGE        := yourdockerhubuser/ulanzi-clock

.PHONY: all build run broker broker-down tidy \
        docker-build-dev docker-push-dev \
        docker-build-prod docker-push-prod \
        k8s-diff-dev k8s-diff-prod argocd-apply

## Build the clock binary
build:
	go build -o $(BINARY) $(CMD)

## Run without building (go run)
run:
	go run $(CMD)/main.go

## Start the Mosquitto broker via Docker Compose
broker:
	docker compose -f docker/docker-compose.yml up -d
	@echo "Mosquitto running on :1883"

## Stop the broker
broker-down:
	docker compose -f docker/docker-compose.yml down

## Tidy dependencies
tidy:
	go mod tidy

## Full local dev: start broker then run clock
dev: broker run

## Cross-compile for Raspberry Pi (arm64) if you want to run it on a Pi
build-pi:
	GOOS=linux GOARCH=arm64 go build -o $(BINARY)-pi $(CMD)

# ── Docker ────────────────────────────────────────────────────────────────────

## Build and push the :dev image (run this while iterating)
docker-dev:
	docker build -t $(IMAGE):dev .
	docker push $(IMAGE):dev

## Build and push a prod release — pass VERSION= to tag it
## Usage: make docker-prod VERSION=1.0.1
docker-prod:
	docker build -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

# ── Kubernetes / Kustomize ────────────────────────────────────────────────────

## Dry-run: show what Kustomize would generate for dev (no apply)
k8s-diff-dev:
	kubectl diff -k k8s/overlays/dev

## Dry-run: show what Kustomize would generate for prod
k8s-diff-prod:
	kubectl diff -k k8s/overlays/prod

## Register both ArgoCD Applications (one-time setup)
argocd-apply:
	kubectl apply -f argocd/application-dev.yaml
	kubectl apply -f argocd/application-prod.yaml

all: build
