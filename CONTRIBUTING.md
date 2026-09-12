# Como contribuir com o DriveFlow

Guia curto para quem vai mexer no código. O fluxo completo de branches está em
[`docs/fluxo-git.md`](docs/fluxo-git.md).

## Antes de começar

```bash
git clone https://github.com/Engenharia-de-Software-com-Devops/DriveFlow.git
cd DriveFlow
cp .env.example .env
make instalar
make testar          # precisa passar antes de você mudar qualquer coisa
```

Se `make testar` falhar em um clone limpo, isso é um defeito do projeto —
abra uma issue em vez de contornar localmente.

## O ciclo de uma mudança

```bash
git checkout dev && git pull origin dev
git checkout -b feature/assunto-da-mudanca
# ... código e testes ...
make verificar       # gofmt + go vet (com e sem a tag `integracao`)
make testar          # unidade do backend + frontend, sem depender de banco
make testar-integracao   # se você mexeu em SQL, migration ou repositório
git push -u origin feature/assunto-da-mudanca
```

Depois, abra o Pull Request para `dev`.

## Checklist do Pull Request

- [ ] `make testar` passa localmente
- [ ] `make verificar` não aponta nada
- [ ] Mudança de comportamento veio acompanhada de teste, na pasta certa:
      sem banco → `backend/tests/unidade/<camada>/`; com Postgres →
      `backend/tests/integracao/`, com `//go:build integracao` na primeira linha
- [ ] Mudança de schema entrou como nova migration `NNNN_descricao.up.sql`,
      com o `.down.sql` correspondente — **migration já aplicada nunca é editada**
- [ ] README atualizado se o modo de executar ou testar mudou
- [ ] Nenhum segredo, `.env`, dump de banco ou artefato de build no diff
- [ ] A descrição do PR diz **como validar** a mudança

## Regras que não se negociam

1. **Migration é imutável.** Depois que um arquivo de migration foi aplicado
   em qualquer ambiente, ele não é editado: a correção vem em uma nova versão.
2. **Toda migration gerada com apoio de IA é revisada por uma pessoa** antes de
   chegar a produção, conforme decidido no diagnóstico do E1.
3. **Nada de push direto em `main`, `prod` ou `homo`.**
4. **Ninguém aprova o próprio PR.**
5. **Valor monetário é inteiro em centavos**, nunca ponto flutuante.
6. **Toda consulta que toca veículo ou locação filtra por empresa.** É o que
   mantém o isolamento entre os tenants.

## Onde fica cada coisa

| Quero mudar... | Mexo em |
| -------------- | ------- |
| Modelo de domínio, erro de domínio ou regra de tarifa | `backend/internal/entities/` |
| Regra de negócio (conflito, validação, fluxo do contrato) | `backend/internal/usecases/` |
| Rota ou código HTTP | `backend/internal/delivery/http/` |
| Schema do banco | `backend/internal/repository/migrations/` (arquivo novo) |
| Consulta SQL | `backend/internal/repository/postgres_repo.go` |
| Teste sem banco | `backend/tests/unidade/` (espelha a camada) |
| Teste que precisa de banco | `backend/tests/integracao/` (tag `integracao`) |
| Montagem das camadas (injeção de dependência) | `backend/cmd/app/main.go` |
| Tela | `frontend/src/componentes/` |
| Chamada à API | `frontend/src/api.js` |
| Containers | `docker-compose.yml`, `*/Dockerfile` |
