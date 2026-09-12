# Evidência de execução e validação — Encontro 2

Registro do que foi executado, com resultado esperado e resultado obtido,
conforme a Etapa 3 e a Etapa 5 da Atividade 2.

---

## 1. Testes automatizados do backend

**Comando**

```bash
cd backend && go test ./... -count=1
```

**Esperado:** todos os pacotes em `ok`; os testes que dependem do PostgreSQL
são ignorados automaticamente quando `DATABASE_URL` não está definida.

**Obtido**

```
?   	driveflow/backend/cmd/app	[no test files]
?   	driveflow/backend/configs	[no test files]
ok  	driveflow/backend/internal/delivery/http	0.879s
ok  	driveflow/backend/internal/entities	0.506s
ok  	driveflow/backend/internal/repository	0.782s
ok  	driveflow/backend/internal/usecases	0.635s
?   	driveflow/backend/pkg/id	[no test files]
```

---

## 2. Testes de integração contra o PostgreSQL

**Comando**

```bash
cd backend && DATABASE_URL="postgres://driveflow:driveflow@localhost:5432/driveflow?sslmode=disable" \
  go test ./... -count=1
```

**Esperado:** os mesmos pacotes em `ok`, agora com os quatro testes de
PostgreSQL executando de verdade — `internal/repository` passa a levar
centésimos de segundo em vez de milésimos.

**Obtido** (rodada anterior à reorganização do backend em clean architecture;
os nomes de pacote e de teste abaixo são os da estrutura antiga e a evidência
precisa ser refeita na próxima execução com o container `db` no ar)

```
?   	driveflow/backend	[no test files]
ok  	driveflow/backend/internal/api	0.006s
ok  	driveflow/backend/internal/armazenamento	0.101s
ok  	driveflow/backend/internal/locacao	0.005s
```

Casos executados nesta rodada:

```
--- PASS: TestMigracoesEstaoVersionadasEmOrdem
--- PASS: TestAplicarMigracoesEhIdempotente
--- PASS: TestPostgresFluxoCompleto
--- PASS: TestPostgresBloqueiaSobreposicaoNoBanco
--- PASS: TestPostgresIsolaTenants
```

---

## 3. Testes do frontend

**Comando**

```bash
cd frontend && CI=true npm test -- --watchAll=false
```

**Esperado:** as duas suítes em `PASS`.

**Obtido**

```
PASS src/api.test.js
PASS src/App.test.js

Test Suites: 2 passed, 2 total
Tests:       10 passed, 10 total
```

---

## 4. Formatação e análise estática

**Comando**

```bash
cd backend && gofmt -l . && go vet ./...
```

**Esperado e obtido:** saída vazia — nenhum arquivo fora do formato padrão e
nenhum apontamento do `go vet`.

---

## 5. Migrations aplicadas em um banco limpo

**Comando**

```bash
cd backend && DATABASE_URL="postgres://driveflow:driveflow@localhost:5432/driveflow?sslmode=disable" go run ./cmd/app
```

**Esperado:** a API conecta no banco, aplica as três migrations em ordem e
passa a ouvir na porta 8080.

**Obtido**

```json
{"level":"INFO","msg":"conectado ao postgres"}
{"level":"INFO","msg":"migracao aplicada","versao":"0001","nome":"0001_criar_empresas"}
{"level":"INFO","msg":"migracao aplicada","versao":"0002","nome":"0002_criar_veiculos"}
{"level":"INFO","msg":"migracao aplicada","versao":"0003","nome":"0003_criar_locacoes"}
{"level":"INFO","msg":"api ouvindo","porta":"8080","versao":"0.1.0"}
```

Estado da tabela de controle depois da subida:

```
 versao |        nome
--------+---------------------
 0001   | 0001_criar_empresas
 0002   | 0002_criar_veiculos
 0003   | 0003_criar_locacoes
(3 rows)
```

Subir a API uma segunda vez não reaplica nada: a mensagem passa a ser
`schema ja atualizado`.

---

## 6. Fluxo de locação pela API

Sequência executada com `curl` contra a API ligada ao PostgreSQL.

| # | Ação | Esperado | Obtido |
| - | ---- | -------- | ------ |
| 1 | `GET /health` | `200`, status `ok` | `{"status":"ok","versao":"0.1.0"}` |
| 2 | `POST /api/empresas` | `201` com o id da empresa | `201` |
| 3 | `POST .../veiculos` placa `ABC1D23`, tarifa `15000` | `201` com status `disponivel` | `201` |
| 4 | `POST .../locacoes` de 10/03 a 13/03 | `201`, valor previsto `45000` (3 diárias × R$ 150,00) | `201`, `"valor_previsto":45000` |
| 5 | `POST .../locacoes` do mesmo veículo de 12/03 a 16/03 | `409`, período sobreposto | `409`, `"veiculo ja reservado no periodo informado"` |
| 6 | `POST .../veiculos` repetindo a placa `ABC1D23` | `409`, placa duplicada | `409`, `"placa ja cadastrada para esta empresa"` |
| 7 | `POST .../devolucao` em 14/03 (um dia de atraso) | `200`, status `encerrada`, valor final `64500` | `200`, `"valor_final":64500`, `"status":"encerrada"` |

O valor final de `64500` confere com a regra: 45000 previsto + 15000 da diária
extra + 4500 de multa (30% sobre a diária em atraso).

---

## 7. Constraint de exclusão do banco, sem passar pela API

Para comprovar que o conflito de reserva é barrado **também** pelo PostgreSQL —
o caso de duas requisições simultâneas —, os `INSERT` foram feitos direto no
banco, contornando a verificação da aplicação.

**Esperado:** o segundo `INSERT`, sobreposto ao primeiro, é recusado.

**Obtido**

```
INSERT 0 1
INSERT 0 1
INSERT 0 1
ERROR:  conflicting key value violates exclusion constraint "locacoes_sem_sobreposicao"
DETAIL:  Key (veiculo_id, tstzrange(inicio, fim_previsto))=(vei-x, ["2026-05-03","2026-05-08"))
         conflicts with existing key (vei-x, ["2026-05-01","2026-05-05")).
```

---

## 8. Build de produção do frontend

**Comando**

```bash
cd frontend && CI=true npm run build
```

**Esperado:** build concluído sem erro nem aviso tratado como erro.

**Obtido:** `The build folder is ready to be deployed.` — bundle principal em
`build/static/js/` e CSS em `build/static/css/`.

---

## 9. Definição dos containers

**Comando**

```bash
docker compose config --quiet
```

**Esperado e obtido:** saída vazia — a definição dos 3 serviços (`db`, `api`,
`web`) é válida.

> **Pendente de validação pela equipe:** `docker compose up --build` ainda não
> foi executado de ponta a ponta. O ambiente em que esta base foi preparada não
> tinha o serviço do Docker disponível, apenas o cliente. A API, as migrations e
> o frontend foram validados diretamente contra um PostgreSQL 16 real, que é o
> mesmo do container `db`. **Antes da entrega, um integrante precisa rodar o
> passo 1 do roteiro abaixo e anexar o resultado a este documento.**

---

## Roteiro de validação em clone limpo (Etapa 5)

Para o integrante que **não** escreveu o código. Siga apenas o README; se algum
passo não funcionar seguindo só o que está escrito, a instrução é que está
errada e precisa ser corrigida.

```bash
# 1. Clone em uma pasta nova, fora de qualquer cópia existente
git clone https://github.com/Engenharia-de-Software-com-Devops/DriveFlow.git validacao-driveflow
cd validacao-driveflow

# 2. Suba os 3 containers
cp .env.example .env
docker compose up --build

# 3. Em outro terminal, confira que os três estão de pé
docker compose ps          # db, api e web com status healthy/running
curl -s localhost:8080/health

# 4. Abra http://localhost:3000 e execute o roteiro manual do README
#    (cadastrar empresa, cadastrar veículo, reservar, tentar reservar de novo,
#     devolver)

# 5. Rode os testes
make testar
```

### Registro da validação

| Item | Responsável | Data | Resultado |
| ---- | ----------- | ---- | --------- |
| Clone limpo e `docker compose up --build` | | | |
| Roteiro manual no navegador | | | |
| `make testar` no clone limpo | | | |
| Instruções do README suficientes? | | | |

Ajustes que a validação exigir entram como `fix/` ou `docs/` a partir de `dev`,
seguindo [`docs/fluxo-git.md`](fluxo-git.md).
