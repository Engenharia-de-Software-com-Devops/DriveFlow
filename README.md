# DriveFlow

Plataforma multi-tenant de locação de veículos para empresas. Cada empresa se
cadastra na plataforma, registra a própria frota e opera o ciclo de locação:
reservar, acompanhar, devolver e encerrar o contrato com o valor calculado.

Projeto integrador da disciplina **Desenvolvimento de Software Integrado —
DevOps** (PG2305-04-Z251, Turma 4 — Z251).

| Integrante | Matrícula |
| ---------- | --------- |
| Mateus     | 2650377   |
| Natan      | 2650295   |
| Jaime      | 2650365   |
| Marcos     | 2651654   |
| Ricardo    | 2650160   |
| Helislana  | 2650139   |
---

## Arquitetura

Três containers, como levantado no diagnóstico do Encontro 1:

```
                 ┌──────────────────┐
  navegador ───► │  web   (React)   │  nginx :8080 → publicado em :3000
                 │  nginx + bundle  │
                 └────────┬─────────┘
                          │  /api, /health (proxy interno)
                 ┌────────▼─────────┐
                 │  api   (Go)      │  :8080
                 │  regras + HTTP   │
                 └────────┬─────────┘
                          │  SQL
                 ┌────────▼─────────┐
                 │  db  (Postgres)  │  :5432
                 │  schema versionado│
                 └──────────────────┘
```

| Camada | Tecnologia | Pasta |
| ------ | ---------- | ----- |
| `web`  | React 19 + TypeScript (Vite), servido por nginx | `frontend/` |
| `api`  | Go 1.24, biblioteca padrão + `lib/pq` | `backend/` |
| `db`   | PostgreSQL 16, migrations versionadas | `backend/internal/repository/migrations/` |

---

## Pré-requisitos

Escolha **um** dos caminhos:

- **Com containers:** Docker 24+ com o plugin `docker compose`.
- **Sem containers:** [Go](https://go.dev/dl/) 1.24+ e [Node.js](https://nodejs.org/) 18+ com npm.

---

## Como executar

### Opção A — os 3 containers (recomendada)

```bash
git clone https://github.com/Engenharia-de-Software-com-Devops/DriveFlow.git
cd DriveFlow
docker compose up --build
```

O `.env` é opcional: sem ele, o compose usa os valores de exemplo. Para trocar
senha ou portas, `cp .env.example .env` e edite (ver
[Variáveis de ambiente](#variáveis-de-ambiente)).

Quando os três containers estiverem no ar:

- Aplicação: <http://localhost:3000>
- API: <http://localhost:8080/health>

Para parar: `docker compose down` (acrescente `-v` para apagar também os dados
do banco).

### A partir das imagens publicadas (Docker Hub)

O CD publica as imagens já validadas pelo CI a cada merge na `main`, no
repositório público
[`jaimegdj/driveflow`](https://hub.docker.com/r/jaimegdj/driveflow). As duas
imagens ficam no mesmo repositório, com o serviço no prefixo da tag:
`api-latest` e `web-latest`, mais `api-<sha>` e `web-<sha>` para cada commit.
Nada é buildado localmente:

```bash
docker pull jaimegdj/driveflow:api-latest
docker pull jaimegdj/driveflow:web-latest
docker compose -f docker-compose.prod.yml up -d
```

Aplicação em <http://localhost:3000>. Para rodar uma versão específica (ou
voltar para uma anterior), passe o SHA do commit:

```bash
DRIVEFLOW_TAG=<sha-do-commit> docker compose -f docker-compose.prod.yml up -d
```

Sem Docker Hub, as mesmas imagens estão no artefato `imagens-docker` de cada
execução do workflow (aba *Actions*): baixe o zip, extraia e rode
`gunzip -c imagens.tar.gz | docker load`.

### Opção B — execução local, sem Docker

Dois terminais.

**Terminal 1 — API** (sobe com armazenamento em memória, sem precisar de banco):

```bash
cd backend
go run ./cmd/app
# api ouvindo na porta 8080
```

Para usar um PostgreSQL de verdade, defina `DATABASE_URL` antes de subir. A API
aplica as migrations pendentes sozinha na inicialização:

```bash
DATABASE_URL="postgres://driveflow:driveflow@localhost:5432/driveflow?sslmode=disable" go run ./cmd/app
```

**Terminal 2 — frontend:**

```bash
cd frontend
npm ci
npm run dev
# abre http://localhost:3000
```

### Desenvolvimento com hot reload nos containers

```bash
docker compose -f docker-compose.dev.yml up --build
```

Defina `LOCAL_UID` e `LOCAL_GID` no `.env` para manter a posse correta dos arquivos
gerados pelos bind mounts. Use `id -u` e `id -g` para obter os valores.

Os módulos Node ficam em `frontend/node_modules`, para que o LSP local resolva
imports e tipos. Os caches Go permanecem no filesystem interno do container,
pois não são necessários para o LSP executado no host.

### Opção C — só o banco em container, API e frontend locais

Três terminais. Útil para desenvolver a API/frontend com hot-reload e
persistência real, sem buildar as imagens de `api` e `web`.

**Terminal 1 — banco:**

```bash
docker compose up -d db
# aguarde ficar "healthy": docker compose ps db
```

**Terminal 2 — API:**

```bash
cd backend
export DATABASE_URL="postgres://driveflow:driveflow@localhost:5432/driveflow?sslmode=disable"
go run ./cmd/app
# api ouvindo na porta 8080
```

**Terminal 3 — frontend:**

```bash
cd frontend
npm ci
npm run dev
# abre http://localhost:3000
```

Para parar só o banco: `docker compose stop db`. Se a porta 5432 já estiver em
uso por outro container/serviço no host, libere-a antes de subir (`docker
compose up -d db` falha com `port is already allocated` nesse caso).

---

## Como testar

A suíte está dividida em dois níveis, e a diferença é só uma: precisa de banco
no ar ou não.

| Alvo | O que roda | Precisa de Docker? |
| ---- | ---------- | ------------------ |
| `make testar` | unidade do backend + frontend | não |
| `make testar-backend` | unidade do backend | não |
| `make testar-frontend` | suítes do React | não |
| `make testar-integracao` | unidade + integração com o Postgres | sim |

```bash
make testar              # o de todo dia: rápido e sem dependência externa
make testar-integracao   # sobe o container db e roda também os testes de banco
```

Sem `make`:

```bash
cd backend  && go test ./... -count=1   # unidade
cd frontend && npm test -- --run        # frontend
```

**Resultado esperado:** todos os pacotes Go em `ok` e as duas suítes do
frontend em `PASS`.

### Onde os testes moram

Os testes do backend ficam em `backend/tests/`, espelhando as camadas da
arquitetura e separados pelo critério que muda a forma de rodar — precisa de
banco ou não:

```
backend/tests/
├── apoio/          montagem compartilhada pelos dois níveis
├── unidade/        entities, usecases, repository, delivery — sem banco
└── integracao/     repository — exige PostgreSQL no ar
```

Todo arquivo em `integracao/` começa com `//go:build integracao`. Sem a tag ele
não entra na compilação, então `go test ./...` roda em qualquer máquina, com ou
sem Docker, e não fica escondendo `SKIP` no meio da saída. Com a tag, os testes
exigem `DATABASE_URL` e falham com mensagem clara se ela não estiver definida.

O detalhamento — em que pasta entra cada tipo de teste novo, e as duas
restrições do Go que explicam o formato — está em
[`backend/tests/README.md`](backend/tests/README.md).

`make verificar` roda `go vet` nas duas configurações, para que o código atrás
da tag não fique sem análise estática.

A evidência da execução registrada pela equipe está em
[`docs/validacao-e2.md`](docs/validacao-e2.md).

---

## Pipeline CI/CD

Um único workflow, [`.github/workflows/ci-cd.yml`](.github/workflows/ci-cd.yml),
com as duas etapas separadas. As execuções ficam na aba
[*Actions*](https://github.com/Engenharia-de-Software-com-Devops/DriveFlow/actions)
do GitHub.

Cada etapa só começa quando a anterior passa.

| Etapa | Job | O que faz | Quando roda |
| ----- | --- | --------- | ----------- |
| CI | 1. Análise estática | `gofmt`, `go vet`, `staticcheck`, direção das dependências entre camadas, migrations, ESLint, `tsc` e validade dos arquivos de compose | todo push e PR para `dev` e `main` |
| CI | 2. Segurança | `govulncheck` e `npm audit` | idem |
| CI | 3. Testes de unidade | Go (`-race`, `-shuffle`, cobertura) e Vitest | idem |
| CI | 4. Testes de integração | testes do repositório contra o Postgres do compose | idem |
| CI | 5. Build | binário da api e bundle do frontend | idem |
| CI | 6. Smoke test | `docker compose up --build --wait`, requisições reais pelo nginx do web e `docker save` das imagens como artefato | idem |
| CD | 7. Publicar | carrega as imagens do smoke test (`docker load`), marca como `<serviço>-<sha>` e `<serviço>-latest` e faz `docker push` em `jaimegdj/driveflow` | só em push na `main`, depois do CI verde |

O CD publica exatamente as imagens que o CI testou: não há rebuild. As
credenciais ficam nos secrets `DOCKERHUB_USERNAME` e `DOCKERHUB_TOKEN`
(*Settings → Secrets and variables → Actions*), nunca no YAML.

`main` e `dev` exigem os jobs de CI verdes antes do merge. Evidência de
pipeline verde, de uma falha real já corrigida e do teste que trava a
entrega se a regra de conflito de reserva quebrar está em
[`docs/validacao-e3.md`](docs/validacao-e3.md).

---

## Variáveis de ambiente

Todas têm valor padrão; o `.env` (fora do git) só é necessário para mudar
algum. Modelo em [`.env.example`](.env.example).

| Variável | Para que serve | Exemplo |
| -------- | -------------- | ------- |
| `POSTGRES_USER` | usuário do banco | `driveflow` |
| `POSTGRES_PASSWORD` | senha do banco | `troque-esta-senha` |
| `POSTGRES_DB` | nome do banco | `driveflow` |
| `POSTGRES_PORT` | porta do banco publicada no host (só `docker-compose.yml`) | `5432` |
| `API_PORT` | porta da api | `8080` |
| `WEB_PORT` | porta do frontend no host | `3000` |
| `DATABASE_URL` | conexão da api com o Postgres; vazia = armazenamento em memória | `postgres://driveflow:senha@db:5432/driveflow?sslmode=disable` |
| `VITE_API_URL` | *build arg* do frontend; vazio = chama `/api` na mesma origem, via nginx | `""` |
| `DRIVEFLOW_TAG` | tag das imagens no `docker-compose.prod.yml` | `latest` ou SHA do commit |
| `LOCAL_UID` / `LOCAL_GID` | dono dos arquivos nos bind mounts do `docker-compose.dev.yml` | `1000` |

---

## Uso de IA

O registro de prompts, respostas, decisões da equipe e evidência de validação
está em [`docs/uso-de-ia.md`](docs/uso-de-ia.md). Toda saída de IA foi tratada
como hipótese até passar por teste, execução ou revisão humana.

---

## Troubleshooting

- **`port is already allocated` ao subir o compose.** Outro serviço usa a
  porta no host (comum: um Postgres local na 5432). Troque só a porta
  publicada, sem mexer no resto: `POSTGRES_PORT=5433 docker compose up --build`
  (o mesmo vale para `API_PORT` e `WEB_PORT`).
- **api reinicia com `banco indisponivel apos 45s: dial tcp: lookup db ...`.**
  A api não acha o serviço `db` na rede do compose, em geral por um container
  `driveflow-db` antigo criado com outra configuração. Recrie os containers
  (os dados ficam no volume): `docker compose up -d --build --force-recreate`.
- **`docker-compose.prod.yml` sobe uma versão antiga.** A tag `latest` já
  estava em cache local. Atualize antes: `docker compose -f
  docker-compose.prod.yml pull`.

---

## Roteiro de verificação manual

Com a aplicação no ar em <http://localhost:3000>:

1. **Cadastre a empresa** — nome `Locadora Alfa`, CNPJ `12345678000190`.
2. **Cadastre um veículo** — placa `ABC1D23`, modelo `Onix 1.0`, categoria
   `Econômico`, tarifa diária `150,00`.
3. **Crie uma reserva** — do dia 10 ao dia 13 do mês seguinte. O contrato
   aparece com valor previsto de **R$ 450,00** (3 diárias).
4. **Tente reservar o mesmo veículo** de 12 a 16. A tela deve recusar com
   *"Reserva recusada: veículo já reservado no período informado"*. Este é o
   comportamento que o diagnóstico do E1 apontou como incidente recorrente.
5. **Registre a devolução** do primeiro contrato. O contrato passa a
   `encerrada` e o valor final é exibido.

O mesmo fluxo pela API:

```bash
curl -s localhost:8080/health

EMPRESA=$(curl -s -X POST localhost:8080/api/empresas \
  -d '{"nome":"Locadora Alfa","cnpj":"12345678000190"}' | jq -r .id)

VEICULO=$(curl -s -X POST localhost:8080/api/empresas/$EMPRESA/veiculos \
  -d '{"placa":"ABC1D23","modelo":"Onix 1.0","categoria":"economico","tarifa_diaria":15000}' | jq -r .id)

# 201 Created
curl -s -X POST localhost:8080/api/empresas/$EMPRESA/locacoes \
  -d "{\"veiculo_id\":\"$VEICULO\",\"cliente\":\"Cliente A\",\"inicio\":\"2026-03-10T10:00:00Z\",\"fim_previsto\":\"2026-03-13T10:00:00Z\"}"

# 409 Conflict — período sobreposto
curl -s -X POST localhost:8080/api/empresas/$EMPRESA/locacoes \
  -d "{\"veiculo_id\":\"$VEICULO\",\"cliente\":\"Cliente B\",\"inicio\":\"2026-03-12T10:00:00Z\",\"fim_previsto\":\"2026-03-16T10:00:00Z\"}"
```

---

## API

Todos os corpos são JSON. Os valores monetários trafegam **em centavos**
(`15000` = R$ 150,00) para não acumular erro de arredondamento.

| Método | Rota | Descrição |
| ------ | ---- | --------- |
| `GET`  | `/health` | Estado e versão da API |
| `POST` | `/api/empresas` | Cadastra uma empresa (tenant) |
| `GET`  | `/api/empresas/{id}/veiculos` | Lista a frota da empresa |
| `POST` | `/api/empresas/{id}/veiculos` | Adiciona um veículo à frota |
| `GET`  | `/api/empresas/{id}/locacoes` | Lista os contratos da empresa |
| `POST` | `/api/empresas/{id}/locacoes` | Cria uma reserva |
| `POST` | `/api/empresas/{id}/locacoes/{locacaoId}/devolucao` | Encerra o contrato |

Códigos de erro:

| Status | Significado |
| ------ | ----------- |
| `404` | Empresa, veículo ou locação inexistente — inclui o caso de uma empresa tentar acessar recurso de outra |
| `409` | Conflito de reserva, placa duplicada ou locação já encerrada |
| `422` | Dados inválidos (CNPJ, placa, período, tarifa) |

---

## Regras de negócio

- **Conflito de reserva.** Um veículo não pode ter dois contratos abertos com
  períodos sobrepostos. A regra é verificada na camada de domínio **e** imposta
  pelo banco, com uma constraint de exclusão sobre
  `(veiculo_id, tstzrange(inicio, fim_previsto))`, o que cobre até a corrida
  entre duas requisições simultâneas.
- **Isolamento entre empresas.** Toda busca de veículo e de locação filtra pela
  empresa. Uma empresa nunca enxerga nem reserva a frota de outra.
- **Cálculo de tarifa.** Qualquer fração de dia conta como diária cheia, com o
  mínimo de uma diária. Devolução em atraso cobra as diárias extras acrescidas
  de 30% de multa; devolução antecipada não reduz o valor contratado.
- **Validações de cadastro.** CNPJ com 14 dígitos, placa no padrão antigo
  (`ABC1234`) ou Mercosul (`ABC1D23`), tarifa maior que zero e placa única por
  empresa.

---

## Estrutura do repositório

```
DriveFlow/
├── backend/                  API em Go (clean architecture)
│   ├── cmd/app/main.go       monta as camadas e sobe o servidor
│   ├── configs/              leitura da configuração de ambiente
│   ├── pkg/                  utilitários genéricos (geração de id)
│   ├── internal/
│   │   ├── entities/         domínio: modelos, erros e regras de tarifa
│   │   ├── usecases/         regras de negócio e a porta do repositório
│   │   ├── repository/       repositório em memória e PostgreSQL
│   │   │   └── migrations/   schema versionado (NNNN_descricao.up.sql)
│   │   └── delivery/http/    rotas HTTP e tradução de erros
│   └── tests/                testes, espelhando as camadas
│       ├── apoio/            montagem compartilhada pelos dois níveis
│       ├── unidade/          sem banco (entra no `make testar`)
│       └── integracao/       exige PostgreSQL (tag `integracao`)
├── frontend/                 SPA em React
│   ├── docker/nginx.conf     serve a SPA e repassa /api para a API
│   └── src/
│       ├── api.ts            cliente HTTP
│       └── componentes/      cadastro de empresa, frota e locações
├── docs/                     diagnóstico, fluxo de branches, evidências
├── docker-compose.yml        os 3 containers, buildados do código
├── docker-compose.prod.yml   os 3 containers, a partir do Docker Hub
└── Makefile                  atalhos de execução e teste
```

---

## Fluxo de trabalho

O repositório usa GitFlow com quatro branches de longa duração:
`main`, `prod`, `homo` e `dev`. Todo trabalho novo nasce em uma branch
`feature/*` a partir de `dev` e volta por Pull Request.

Detalhes em [`docs/fluxo-git.md`](docs/fluxo-git.md) e
[`CONTRIBUTING.md`](CONTRIBUTING.md).

---

## Rastreamento do diagnóstico (E1)

Cada gargalo levantado no Encontro 1 tem endereço no código:

| Gargalo (E1) | Onde foi endereçado nesta base |
| ------------ | ------------------------------ |
| Build e deploy manual, sem migrations versionadas | `docker-compose.yml`, `backend/Dockerfile`, `frontend/Dockerfile`, `backend/internal/repository/migrations/` |
| Ausência de testes automatizados entre api, web e db | 48 casos de teste automatizados: domínio, API HTTP, integração com Postgres e interface React |
| Sem observabilidade compartilhada | Log estruturado em JSON na API e `HEALTHCHECK` nos containers |

O diagnóstico completo e o rastreio detalhado estão em
[`docs/diagnostico-e1.md`](docs/diagnostico-e1.md).

**Encontro 3:** `make testar` virou pipeline de integração contínua
([`ci-cd.yml`](.github/workflows/ci-cd.yml)), com status check obrigatório antes do
merge em `dev` e `main`. Evidência em
[`docs/validacao-e3.md`](docs/validacao-e3.md).
