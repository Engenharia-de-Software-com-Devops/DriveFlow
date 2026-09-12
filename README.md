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

---

## Arquitetura

Três containers, como levantado no diagnóstico do Encontro 1:

```
                 ┌──────────────────┐
  navegador ───► │  web   (React)   │  nginx :80  → publicado em :3000
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
| `web`  | React 19 (Create React App), servido por nginx | `frontend/` |
| `api`  | Go 1.24, biblioteca padrão + `lib/pq` | `backend/` |
| `db`   | PostgreSQL 16, migrations versionadas | `backend/internal/armazenamento/migracoes/` |

---

## Pré-requisitos

Escolha **um** dos caminhos:

- **Com containers:** Docker 24+ com o plugin `docker compose`.
- **Sem containers:** [Go](https://go.dev/dl/) 1.24+ e [Node.js](https://nodejs.org/) 18+ com npm.

---

## Como executar

### Opção A — os 3 containers (recomendada)

```bash
cp .env.example .env      # ajuste a senha do banco antes de subir
docker compose up --build
```

Quando os três containers estiverem no ar:

- Aplicação: <http://localhost:3000>
- API: <http://localhost:8080/health>

Para parar: `docker compose down` (acrescente `-v` para apagar também os dados
do banco).

### Opção B — execução local, sem Docker

Dois terminais.

**Terminal 1 — API** (sobe com armazenamento em memória, sem precisar de banco):

```bash
cd backend
go run .
# api ouvindo na porta 8080
```

Para usar um PostgreSQL de verdade, defina `DATABASE_URL` antes de subir. A API
aplica as migrations pendentes sozinha na inicialização:

```bash
DATABASE_URL="postgres://driveflow:driveflow@localhost:5432/driveflow?sslmode=disable" go run .
```

**Terminal 2 — frontend:**

```bash
cd frontend
npm install
npm start
# abre http://localhost:3000
```

---

## Como testar

```bash
make testar              # backend + frontend
make testar-backend      # go test ./...
make testar-frontend     # CI=true npm test
make testar-integracao   # sobe o container db e roda os testes contra o Postgres
```

Sem `make`:

```bash
cd backend  && go test ./... -count=1
cd frontend && CI=true npm test -- --watchAll=false
```

**Resultado esperado:** todos os pacotes Go em `ok` e as duas suítes do
frontend em `PASS`. Os testes que dependem do Postgres são ignorados
automaticamente (`SKIP`) quando `DATABASE_URL` não está definida, então
`go test ./...` funciona em qualquer máquina, com ou sem Docker.

A evidência da execução registrada pela equipe está em
[`docs/validacao-e2.md`](docs/validacao-e2.md).

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
├── backend/                  API em Go
│   ├── main.go               escolhe o armazenamento e sobe o servidor
│   └── internal/
│       ├── api/              rotas HTTP e tradução de erros
│       ├── locacao/          domínio: modelo, regras, tarifa
│       └── armazenamento/    repositório em memória e PostgreSQL
│           └── migracoes/    schema versionado (NNNN_descricao.up.sql)
├── frontend/                 SPA em React
│   ├── nginx.conf            serve a SPA e repassa /api para a API
│   └── src/
│       ├── api.js            cliente HTTP
│       └── componentes/      cadastro de empresa, frota e locações
├── docs/                     diagnóstico, fluxo de branches, evidências
├── docker-compose.yml        os 3 containers
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
| Build e deploy manual, sem migrations versionadas | `docker-compose.yml`, `backend/Dockerfile`, `frontend/Dockerfile`, `backend/internal/armazenamento/migracoes/` |
| Ausência de testes automatizados entre api, web e db | 48 casos de teste automatizados: domínio, API HTTP, integração com Postgres e interface React |
| Sem observabilidade compartilhada | Log estruturado em JSON na API e `HEALTHCHECK` nos containers |

O diagnóstico completo e o rastreio detalhado estão em
[`docs/diagnostico-e1.md`](docs/diagnostico-e1.md).

**Próximo incremento (E3):** transformar `make testar` em pipeline de
integração contínua, com status check obrigatório antes do merge.
