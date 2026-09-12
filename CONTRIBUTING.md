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
make verificar       # gofmt + go vet
make testar
git push -u origin feature/assunto-da-mudanca
```

Depois, abra o Pull Request para `dev`.

## Checklist do Pull Request

- [ ] `make testar` passa localmente
- [ ] `make verificar` não aponta nada
- [ ] Mudança de comportamento veio acompanhada de teste
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
| Regra de negócio (tarifa, conflito, validação) | `backend/internal/locacao/` |
| Rota ou código HTTP | `backend/internal/api/` |
| Schema do banco | `backend/internal/armazenamento/migracoes/` (arquivo novo) |
| Consulta SQL | `backend/internal/armazenamento/postgres.go` |
| Tela | `frontend/src/componentes/` |
| Chamada à API | `frontend/src/api.js` |
| Containers | `docker-compose.yml`, `*/Dockerfile` |
