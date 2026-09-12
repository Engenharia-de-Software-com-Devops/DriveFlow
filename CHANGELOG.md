# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/).
Versionamento conforme [SemVer](https://semver.org/lang/pt-BR/).

## [0.1.0] — 2026-09-12

Primeira base executável do projeto integrador (Encontro 2).

### Adicionado

**Domínio de locação (`api`)**
- Cadastro de empresas (tenants), com validação de CNPJ.
- Cadastro de frota, com placa no padrão antigo e Mercosul, única por empresa.
- Criação de reserva com bloqueio de períodos sobrepostos para o mesmo veículo.
- Devolução com cálculo de valor final, incluindo multa de 30% por diária em atraso.
- Isolamento entre empresas em toda consulta de veículo e de locação.
- API HTTP JSON com tradução de erros de domínio em `404`, `409` e `422`.
- `GET /health` com estado e versão; log estruturado em JSON por requisição.

**Persistência (`db`)**
- Schema do PostgreSQL versionado em migrations `NNNN_descricao.up.sql`, com
  `.down.sql` correspondente.
- Migrations aplicadas na subida da API, cada uma na própria transação junto do
  registro em `schema_migracoes`; reaplicação é inofensiva.
- Constraint de exclusão sobre `(veiculo_id, tstzrange(inicio, fim_previsto))`
  nas locações abertas, impedindo reserva sobreposta mesmo em requisições
  simultâneas.
- Armazenamento em memória como alternativa, usado quando `DATABASE_URL` não
  está definida.

**Interface (`web`)**
- Tela de cadastro de empresa, gestão de frota e operação de locações.
- Exibição do motivo da recusa quando a API devolve conflito de reserva.
- Aviso quando a API está fora do ar.

**Infraestrutura**
- `docker-compose.yml` com os 3 containers (`db`, `api`, `web`), healthcheck e
  dependência ordenada.
- Dockerfiles em dois estágios para API e frontend; nginx servindo a SPA e
  repassando `/api` para a API.
- `Makefile` com os comandos de execução, teste e containers.

**Qualidade**
- 48 casos de teste automatizados: 38 no backend (domínio, API HTTP e
  integração real com PostgreSQL) e 10 no frontend.

**Documentação**
- README com execução, teste e roteiro de verificação manual.
- `CONTRIBUTING.md`, fluxo de branches, rastreio do diagnóstico do E1,
  evidência de execução e registro de uso de IA.

### Conhecido e ainda em aberto

- Sem esteira de CI/CD — tema do Encontro 3.
- Sem autenticação: o tenant é identificado pelo id na URL.
- Sem painel único de observabilidade dos 3 containers.
- `docker compose up --build` pendente de validação de ponta a ponta pela equipe.

[0.1.0]: https://github.com/Engenharia-de-Software-com-Devops/DriveFlow/releases/tag/v0.1.0
