# Validação do pipeline de CI — Encontro 3

Evidência de que o pipeline (`.github/workflows/ci.yml`) protege a entrega:
roda instalação, teste real e build a cada push/PR em `dev` e `main`, e uma
falha real já bloqueou (e foi corrigida) antes de chegar em `dev`.

---

## 1. Workflow versionado

[`​.github/workflows/ci.yml`](../.github/workflows/ci.yml) dispara em `push` e
`pull_request` para `dev` e `main`, com cinco jobs:

| Job | O que roda |
| --- | --- |
| `verificar` | `gofmt`, `go vet` (com e sem a tag `integracao`) |
| `testar-backend` | `go test ./...` |
| `testar-frontend` | `npm ci` + `npm test -- --run` |
| `testar-integracao` | sobe `db` via `docker compose`, roda os testes com a tag `integracao` e `DATABASE_URL` real |
| `build` | `go build ./cmd/app` e `npm run build`, só depois que `testar-backend` e `testar-frontend` passam |

## 2. Pipeline verde

Execução completa com os cinco jobs em `success`, incluindo o build:
<https://github.com/Engenharia-de-Software-com-Devops/DriveFlow/actions/runs/34698115801>

## 3. Falha real, explicada, e a correção aplicada

A migração do frontend de Create React App para Vite (PR #7) trocou o runner
de teste de `react-scripts` para `vitest`, mas o workflow ainda chamava o
teste com a flag do CRA:

- **Falha:** <https://github.com/Engenharia-de-Software-com-Devops/DriveFlow/actions/runs/34697105103>
  — job `Frontend — testes`, comando `npm test -- --watchAll=false`:

  ```
  CACError: Unknown option `--watchAll`
      at Command.checkUnknownOptions (.../vitest/dist/chunks/cac.D805sv8h.js:405:17)
  Node.js v22.23.2
  ##[error]Process completed with exit code 1.
  ```

  `--watchAll` é uma flag do Jest/react-scripts; o `vitest` não a reconhece e
  aborta antes de rodar qualquer teste. O job barrou o merge do PR #7 nesse
  estado.

- **Causa raiz:** `ci.yml` não foi atualizado junto com a troca do test
  runner no mesmo PR.

- **Correção:** commit `a9b937f` (`fix: resolved ci`), trocando o comando
  para `npm test -- --run` (a flag equivalente no `vitest`).

- **Confirmação:** execução seguinte, já verde:
  <https://github.com/Engenharia-de-Software-com-Devops/DriveFlow/actions/runs/34697353633>

## 4. Teste forte: quebra se a regra principal quebrar

A regra central do domínio — um veículo não pode ter dois contratos abertos
com períodos sobrepostos — tem cobertura em duas camadas, as duas rodando no
CI:

- `TestPeriodosSobrepostos` (`backend/tests/unidade/...`) — job
  `testar-backend`, sem banco.
- `TestPostgresBloqueiaSobreposicaoNoBanco` (`backend/tests/integracao/...`)
  — job `testar-integracao`, contra Postgres real, provando que a
  `EXCLUDE CONSTRAINT` do banco também rejeita a sobreposição, não só a
  validação em memória.

Não é um teste trivial de estrutura: ele insere duas locações com intervalos
que se cruzam e verifica que a segunda é recusada (`409`) — se a checagem de
conflito for removida ou quebrada, os dois testes falham e o job correspondente
fica vermelho.

## 5. Falha bloqueia a entrega

`main` e `dev` exigem, via branch protection, que os cinco jobs acima
terminem em sucesso antes de permitir o merge de um Pull Request.

## 6. Registro de IA

Ver [`Registro 5`](uso-de-ia.md#registro-5--auditoria-do-pipeline-de-ci) em
`uso-de-ia.md`.
