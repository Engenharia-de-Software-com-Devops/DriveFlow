# Fluxo de branches (GitFlow)

Este documento define como o código caminha da máquina de quem desenvolve até
produção. Ele vale a partir da base publicada no Encontro 2.

---

## Branches de longa duração

Quatro branches nunca são apagadas. Cada uma representa um estágio da esteira:

| Branch | Papel | Quem escreve nela | Ambiente |
| ------ | ----- | ----------------- | -------- |
| `dev`  | Integração contínua do time. Recebe todas as `feature/*`. | Merge de `feature/*` via PR | Desenvolvimento |
| `homo` | Homologação. Recebe uma `release/*` fechada para validação. | Merge de `release/*` | Homologação / QA |
| `prod` | O que está rodando para as empresas clientes. | Merge de `homo` já aprovado | Produção |
| `main` | Linha oficial e estável. Espelha produção e guarda as tags de versão. | Merge de `prod` | — |

> `dev` é a branch mais movimentada e a única em que se espera conflito com
> frequência. `main` só muda quando uma versão inteira é fechada.

---

## Branches temporárias

| Prefixo | Nasce de | Volta para | Some depois do merge |
| ------- | -------- | ---------- | -------------------- |
| `feature/<assunto>` | `dev` | `dev` | sim |
| `fix/<assunto>` | `dev` | `dev` | sim |
| `docs/<assunto>` | `dev` | `dev` | sim |
| `release/<versao>` | `dev` | `homo` → `prod` → `main` | sim |
| `hotfix/<versao>` | `main` | `main` + `prod` + `dev` | sim |

Nomes em minúsculas, palavras separadas por hífen, sem acento:
`feature/api-locacao`, `fix/calculo-multa-atraso`, `release/0.1.0`.

---

## O caminho normal de uma mudança

```
feature/*  ──┐
fix/*      ──┼──► dev ──► release/0.1.0 ──► homo ──► prod ──► main ──► tag v0.1.0
docs/*     ──┘                                                 ▲
                                                               │
                                          hotfix/0.1.1 ────────┘
                                    (volta também para prod e dev)
```

1. **Abrir a branch** a partir de `dev` atualizada.
2. **Desenvolver e testar localmente** (`make testar` precisa passar).
3. **Abrir o PR** para `dev`, com descrição do que muda e como validar.
4. **Revisão de outro integrante** — ninguém aprova o próprio PR.
5. **Merge em `dev`** e remoção da branch temporária.
6. Quando o conjunto em `dev` estiver fechado, **abrir `release/<versao>`**,
   congelar o escopo e corrigir nela apenas o que a homologação apontar.
7. **`release/*` → `homo`**, validar em homologação.
8. **`homo` → `prod`** após aprovação do responsável técnico.
9. **`prod` → `main`** e criar a tag `vX.Y.Z`.

## Correção urgente em produção

Bug em produção não espera o ciclo completo:

```bash
git checkout main && git pull
git checkout -b hotfix/0.1.1
# corrige, testa, commita
```

O hotfix entra em `main` e `prod` e **é obrigatoriamente propagado de volta
para `dev`**, senão o próximo release reintroduz o bug.

---

## Comandos do dia a dia

```bash
# começar uma feature
git checkout dev
git pull origin dev
git checkout -b feature/relatorio-de-frota

# publicar para abrir o PR
git push -u origin feature/relatorio-de-frota

# manter a feature em dia com dev durante o desenvolvimento
git fetch origin
git merge origin/dev

# fechar uma release
git checkout dev && git pull origin dev
git checkout -b release/0.2.0
git push -u origin release/0.2.0
```

Os merges das branches temporárias usam `--no-ff`, para que o histórico
mostre onde cada bloco de trabalho começou e terminou:

```bash
git merge --no-ff feature/relatorio-de-frota
```

---

## Convenção de mensagens de commit

Formato `tipo(escopo): resumo no imperativo`, resumo em minúsculas e sem ponto
final, com até 72 caracteres.

| Tipo | Uso |
| ---- | --- |
| `feat` | nova funcionalidade |
| `fix` | correção de defeito |
| `docs` | documentação |
| `test` | testes |
| `build` | containers, dependências, empacotamento |
| `chore` | tarefas de manutenção |
| `merge` | integração de branch |

Exemplos:

```
feat(api): expoe o dominio de locacao como API HTTP JSON
fix(tarifa): corrige multa em devolucao no mesmo dia
docs(readme): descreve a execucao dos 3 containers
```

O corpo do commit explica **por quê**, não o que o diff já mostra.

---

## Proteções da `main`

Combinadas pela equipe e aplicadas após a publicação da base:

- `main`, `prod` e `homo` não aceitam push direto — só merge por PR.
- Todo PR precisa da aprovação de ao menos um outro integrante.
- A branch precisa estar atualizada com o destino antes do merge.
- A partir do E3, o status check da integração contínua passa a ser
  obrigatório: PR com teste vermelho não entra.

> **Exceção de inicialização:** o primeiro commit da base do projeto foi
> direto na `main`, como previsto na Atividade 2. A partir dele, toda mudança
> segue por branch e Pull Request.

---

## Histórico desta base

A base publicada no Encontro 2 foi organizada nas seguintes branches:

| Branch | Conteúdo |
| ------ | -------- |
| `feature/api-locacao` | Domínio de locação, regras de tarifa e conflito de reserva; API HTTP |
| `feature/persistencia-postgres` | Migrations versionadas e adaptador PostgreSQL |
| `feature/web-frota` | Frontend React consumindo a API |
| `feature/containers-docker-compose` | Dockerfiles, nginx e os 3 containers |
| `docs/base-do-projeto` | README, este documento, evidências e registro de uso de IA |
| `release/0.1.0` | Fechamento da versão 0.1.0 |
