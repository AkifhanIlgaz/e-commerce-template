.PHONY: dev run build templ infra infra-down

# Mongo + Redis'i docker ile ayağa kaldır
infra:
	docker compose up -d

infra-down:
	docker compose down

# templ dosyalarını Go koduna çevir
templ:
	templ generate

build: templ
	go build -o bin/server ./cmd/server

run: templ
	go run ./cmd/server/main.go

# infra + hot-reload'lu sunucu (air: .go/.templ/.css değişince templ generate + rebuild)
dev: infra
	air
