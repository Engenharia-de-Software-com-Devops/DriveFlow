# Convenções do DriveFlow

Instruções para qualquer pessoa — ou assistente — que for mexer neste
repositório. O fluxo completo está em [`docs/fluxo-git.md`](docs/fluxo-git.md);
aqui ficam as regras que não podem ser esquecidas.

## Nomes de branch

Use **apenas** estes prefixos, em minúsculas, palavras separadas por hífen e
sem acento:

| Prefixo | Para quê | Nasce de |
| ------- | -------- | -------- |
| `feature/` | Funcionalidade nova | `dev` |
| `fix/` | Correção de defeito | `dev` |
| `docs/` | Documentação | `dev` |
| `release/` | Fechamento de versão | `dev` |
| `hotfix/` | Correção urgente em produção | `main` |

Exemplos: `feature/relatorio-de-frota`, `fix/calculo-multa-atraso`,
`release/0.2.0`.

**Nunca nomeie uma branch com `claude`, com o nome de qualquer ferramenta ou
assistente, nem com identificador gerado automaticamente.** O nome da branch
descreve o trabalho, não quem ou o que o executou. O mesmo vale para mensagens
de commit, descrições de PR e comentários no código: nenhum deles cita
ferramenta ou assistente.

`main`, `prod`, `homo` e `dev` são permanentes e nunca recebem push direto.

## Commits

Formato `tipo(escopo): resumo no imperativo`, em minúsculas, sem ponto final,
até 72 caracteres. Tipos: `feat`, `fix`, `docs`, `test`, `build`, `chore`,
`merge`. O corpo explica **por quê**, não o que o diff já mostra.

Merge de branch temporária sempre com `--no-ff`.

## Antes de abrir um PR

```bash
make verificar    # gofmt + go vet, saída precisa ser vazia
make testar       # backend + frontend, tudo verde
```

Checklist completo em [`CONTRIBUTING.md`](CONTRIBUTING.md).

## Regras do código

1. **Migration já aplicada nunca é editada.** A correção vem como um novo
   arquivo `NNNN_descricao.up.sql`, com o `.down.sql` correspondente.
2. **Toda migration é revisada por uma pessoa** antes de chegar a produção.
3. **Valor monetário é inteiro em centavos**, nunca ponto flutuante. A
   conversão para reais acontece só na formatação da tela.
4. **Toda consulta que toca veículo ou locação filtra por empresa.** É o que
   mantém o isolamento entre os tenants.
5. **Mudança de comportamento vem com teste.** Regra de negócio nova sem teste
   não entra.
6. **Nada de segredo, `.env`, dump de banco ou artefato de build no diff.**

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

## Documentação em texto

Português do Brasil, com acentuação correta em arquivos `.md`. Código,
identificadores, mensagens de log e comentários seguem sem acento, como o
restante do projeto já faz.
