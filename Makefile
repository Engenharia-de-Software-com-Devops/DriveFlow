# Atalhos para os comandos do projeto. `make` sem argumento lista os alvos.
.DEFAULT_GOAL := ajuda
.PHONY: ajuda instalar testar testar-backend testar-frontend testar-integracao \
        verificar api web subir derrubar logs limpar

ajuda: ## Lista os alvos disponiveis
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

instalar: ## Baixa as dependencias do backend e do frontend
	cd backend && go mod download
	cd frontend && npm install

testar: testar-backend testar-frontend ## Roda todos os testes automatizados

testar-backend: ## Testes do backend (os de banco sao ignorados sem DATABASE_URL)
	cd backend && go test ./... -count=1

testar-frontend: ## Testes do frontend
	cd frontend && CI=true npm test -- --watchAll=false

testar-integracao: ## Testes de integracao contra o postgres do compose
	docker compose up -d db
	cd backend && DATABASE_URL="postgres://driveflow:driveflow@localhost:5432/driveflow?sslmode=disable" \
		go test ./... -count=1 -v

verificar: ## Formatacao e analise estatica do backend
	cd backend && gofmt -l . && go vet ./...

api: ## Sobe a api local (em memoria, sem precisar de banco)
	cd backend && go run ./cmd/app

web: ## Sobe o frontend local em modo desenvolvimento
	cd frontend && npm start

subir: ## Sobe os 3 containers (db, api, web)
	docker compose up --build -d
	@echo "Aplicacao em http://localhost:3000 | API em http://localhost:8080/health"

derrubar: ## Para os 3 containers
	docker compose down

logs: ## Acompanha os logs dos 3 containers
	docker compose logs -f

limpar: ## Remove containers, volumes e artefatos de build
	docker compose down -v
	rm -rf frontend/build frontend/coverage backend/bin
