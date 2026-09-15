# Do diagnóstico (E1) à base executável (E2)

Este documento liga o diagnóstico DevOps do Encontro 1 ao código publicado no
Encontro 2. Ele existe para responder a uma pergunta: **o que da base entregue
atende ao que a equipe apontou como problema?**

---

## Resumo do diagnóstico

**Projeto:** plataforma multi-tenant de locação de veículos para empresas.
Cada empresa cadastra a própria frota e opera criação de reserva, retirada,
devolução, cálculo de valor e encerramento de contrato. Arquitetura em 3
containers: `api` (Go), `web` (React) e `db` (PostgreSQL).

### Fluxo mapeado na Fase 1

| Etapa | Situação encontrada |
| ----- | ------------------- |
| Pedido/demanda | Planilha compartilhada, sem backlog priorizado nem critério de aceite |
| Desenvolvimento | Containers locais, mas schema alterado à mão no Postgres, com `.sql` soltos |
| Teste | Manual no navegador e no Postman; sem teste automatizado cobrindo api + web + db |
| Entrega | Imagens reconstruídas na máquina do desenvolvedor, enviadas por SCP/SSH, sem rollback |
| Operação | Monitoramento reativo; falhas descobertas quando o cliente reclama |

### Gargalos classificados na Fase 2

1. **Ferramenta/dado** — build e deploy manuais dos 3 containers, sem esteira de
   CI/CD e sem migrations versionadas.
2. **Processo** — ausência de testes automatizados validando a integração entre
   `api`, `web` e `db`.
3. **Cultura** — sem dono formal da operação nem observabilidade compartilhada.

### Melhoria priorizada na Fase 3

Criar uma esteira de CI/CD que construa e teste os 3 containers a cada commit,
incluindo migrations versionadas e testes automatizados do fluxo de cadastro de
frota e locação, antes de qualquer deploy.

---

## O que o E2 entregou para cada gargalo

### Gargalo 1 — build e deploy manual, sem migrations versionadas

| O que foi feito | Onde |
| --------------- | ---- |
| Os 3 containers sobem com um comando, com healthcheck e ordem de dependência | `docker-compose.yml` |
| Build reproduzível: binário estático em dois estágios; `npm ci` fixado no lockfile | `backend/Dockerfile`, `frontend/Dockerfile` |
| Schema versionado em arquivos `NNNN_descricao.up.sql`, com `.down.sql` para rollback | `backend/internal/repository/migrations/` |
| Migrations aplicadas automaticamente na subida da API, cada uma na própria transação junto do registro em `schema_migracoes` | `backend/internal/repository/migration.go` |
| Aplicação idempotente: reiniciar o container não reaplica nada | `TestApplyMigrationsIsIdempotent` |

Efeito prático: o passo "aplicar o `.sql` na mão no servidor" deixou de existir.
Uma migration fora de ordem não tem mais como acontecer, porque a ordem é o
próprio nome do arquivo e o que já foi aplicado está registrado no banco.

### Gargalo 2 — ausência de testes automatizados

O diagnóstico citou três regras críticas que só eram descobertas em produção.
As três viraram teste:

| Regra crítica (E1) | Teste |
| ------------------ | ----- |
| Conflito de reserva entre empresas | `TestReservarVeiculoRejeitaPeriodoSobreposto`, `TestReservaConflitanteRetorna409`, `TestPostgresBloqueiaSobreposicaoNoBanco` |
| Cálculo correto de tarifa | `TestValorFinalComAtraso`, `TestDiarias`, `TestDevolverVeiculoCalculaValorFinalComMulta` |
| Disponibilidade real da frota | `TestDevolverVeiculoLiberaParaNovaReserva`, `TestListarFrotaIsolaPorEmpresa`, `TestEmpresaNaoReservaVeiculoDeOutraEmpresa` |

Distribuição dos 48 casos automatizados:

| Camada | Casos | O que cobre |
| ------ | ----- | ----------- |
| `backend/internal/entities` | 16 | Regras de tarifa e sobreposição de períodos |
| `backend/internal/usecases` | 11 | Conflito de reserva, isolamento entre tenants, ciclo do contrato |
| `backend/internal/delivery/http` | 5 | Rotas HTTP e tradução de erro de domínio em status |
| `backend/internal/repository` | 6 | Migrations versionadas e integração real com o PostgreSQL |
| `frontend/src` | 10 | Cliente HTTP e comportamento da interface |

O conflito de reserva é barrado em **duas** camadas: na regra de domínio e na
constraint de exclusão do PostgreSQL. A segunda cobre o caso que o time não
conseguia reproduzir manualmente — duas requisições simultâneas passando pela
verificação ao mesmo tempo.

### Gargalo 3 — sem observabilidade compartilhada

Este era o gargalo cultural e o E2 entrega apenas a fundação técnica dele:

| O que foi feito | Onde |
| --------------- | ---- |
| Log estruturado em JSON, com método, rota e duração de cada requisição | `backend/cmd/app/main.go`, `backend/internal/delivery/http/server.go` |
| `HEALTHCHECK` nos containers `api` e `web`, e `pg_isready` no `db` | `*/Dockerfile`, `docker-compose.yml` |
| `GET /health` reportando estado e versão | `backend/internal/delivery/http/server.go` |
| Conhecimento do fluxo fora da cabeça de uma pessoa: README, `CONTRIBUTING.md` e fluxo de branches documentado | `README.md`, `CONTRIBUTING.md`, `docs/fluxo-git.md` |

O que **ainda falta**: painel único com as métricas dos 3 containers e dono
formal da operação. Continua em aberto para os próximos encontros.

---

## O que o E2 deliberadamente não fez

Manter honesto o que ficou de fora é parte da entrega:

- **Sem esteira de CI/CD.** A melhoria priorizada na Fase 3 é justamente o tema
  do Encontro 3. O E2 entrega a validação executável localmente (`make testar`);
  o E3 a automatiza e a transforma em status check obrigatório.
- **Sem deploy automatizado nem plano de rollback.** O `.down.sql` de cada
  migration prepara o terreno, mas o procedimento de rollback ainda não existe.
- **Sem autenticação.** A empresa é identificada pelo id na URL. Para uma
  plataforma multi-tenant real isso é insuficiente e está registrado como
  dívida técnica conhecida.
- **Sem backlog formal.** O gargalo do "pedido/demanda" por planilha não foi
  tratado nesta entrega.

---

## Evidência de sucesso proposta no E1

O diagnóstico definiu duas medidas. Elas só ficam mensuráveis a partir do E3,
mas a base do E2 já cria a linha de partida:

| Medida do E1 | Situação |
| ------------ | -------- |
| Incidentes pós-deploy por conflito de reserva ou tarifa em 30 dias | As duas regras passam a ter teste automatizado; incidentes por elas devem cair a zero na faixa coberta |
| Tempo médio de build + deploy dos 3 containers | Passível de medição assim que a esteira do E3 existir; hoje o build local dos 3 containers é um comando só |
