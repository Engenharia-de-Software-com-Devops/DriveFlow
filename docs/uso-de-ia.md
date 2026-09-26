# Registro de uso de IA — Encontro 2

A Atividade 2 pede o registro de **prompt, resposta, decisão e evidência de
validação**. O diagnóstico do E1 já havia definido onde a IA poderia ajudar
(gerar testes de integração, sugerir migrations, triar logs) e o que exige
validação humana obrigatória (revisão de qualquer migration antes de produção,
aprovação da esteira pelo responsável técnico e conferência manual das regras
de locação).

Cada bloco abaixo segue o mesmo formato.

---

## Registro 1 — Testes das regras de tarifa e conflito de reserva

**Prompt**
> Com base nestas regras de locação (diária cheia para qualquer fração de dia,
> multa de 30% sobre diária em atraso, um veículo não pode ter contratos
> abertos com períodos sobrepostos), quais casos faltam para os limites e os
> erros dessas funções? Explique o que cada teste comprova.

**Resposta (resumo)**
Sugeriu casos de limite que o time não tinha listado: período de duração zero
ou invertido, devolução exatamente no horário previsto, devolução antecipada e
reservas cujos limites apenas se encostam (fim de uma igual ao início da outra).

**Decisão da equipe**
Aceitos. O caso dos limites encostados virou regra explícita: intervalo é
fechado no início e aberto no fim, `[inicio, fim)`, então uma reserva que
começa exatamente quando a outra termina é permitida. A sugestão de aceitar
período invertido foi **recusada**: devolve `422`, porque data invertida é erro
de digitação do atendente, não um contrato de duração mínima.

**Evidência de validação**
`TestDiarias`, `TestPeriodosSobrepostos` e `TestReservarVeiculoAceitaPeriodoSeguinte`
passando — item 1 de [`validacao-e2.md`](validacao-e2.md).

---

## Registro 2 — Migration das locações

**Prompt**
> Sugira o script de migration versionada para a tabela de locações, sabendo
> que o mesmo veículo não pode ter dois contratos abertos com períodos
> sobrepostos.

**Resposta (resumo)**
Sugeriu a tabela com chaves estrangeiras e uma constraint
`EXCLUDE USING gist (veiculo_id WITH =, tstzrange(inicio, fim_previsto) WITH &&)`,
com a extensão `btree_gist`.

**Decisão da equipe**
Aceita **com alterações**, e apenas depois de uma pessoa revisar o SQL linha a
linha, como o E1 exige. Acrescentamos `WHERE (status = 'aberta')` à constraint,
porque a sugestão original bloqueava períodos sobrepostos mesmo entre contratos
já encerrados — o que impediria relocar um veículo devolvido antes do prazo. A
constraint **não** substitui a verificação na camada de domínio: a aplicação
precisa devolver `409` com mensagem legível, e não um erro do banco.

**Evidência de validação**
`TestDevolverVeiculoLiberaParaNovaReserva` (o caso que a sugestão original
quebrava), `TestPostgresBloqueiaSobreposicaoNoBanco` e os `INSERT` diretos no
banco — itens 2 e 7 de [`validacao-e2.md`](validacao-e2.md).

---

## Registro 3 — Revisão do diff antes de publicar

**Prompt**
> Com estes critérios de aceite da Atividade 2, encontre bugs e regressões no
> diff. Cite a linha e uma entrada de reprodução.

**Resposta (resumo)**
Apontou dois problemas concretos: valores monetários em ponto flutuante
acumulariam erro de arredondamento em contratos longos, e o `README` descrevia
um comando de teste que falhava em máquina sem PostgreSQL.

**Decisão da equipe**
Os dois aceitos. Valores passaram a ser inteiros em centavos em toda a pilha,
convertidos para reais só na formatação da tela. Os testes de banco passaram a
ser ignorados automaticamente sem `DATABASE_URL`, para que `go test ./...`
funcione em qualquer máquina.

**Evidência de validação**
`TestValorFinalComAtraso` conferindo `64500` em centavos, e `go test ./...`
passando sem `DATABASE_URL` — itens 1 e 6 de [`validacao-e2.md`](validacao-e2.md).

---

## Registro 4 — Avaliação de risco da base

**Prompt**
> Que hipótese sua depende de contexto ausente? Como validá-la antes de alterar
> o código?

**Resposta (resumo)**
Listou três hipóteses assumidas sem confirmação: que a tarifa é por diária e
não por hora; que a devolução antecipada não gera reembolso; e que a empresa
seria identificada apenas pelo id na URL, sem autenticação.

**Decisão da equipe**
As duas primeiras foram confirmadas como regra do projeto e estão documentadas
no README. A terceira foi registrada como **dívida técnica conhecida** em
[`diagnostico-e1.md`](diagnostico-e1.md): identificar o tenant por id na URL
não é suficiente para uma plataforma multi-tenant real e precisa de
autenticação em um incremento futuro.

**Evidência de validação**
Regras descritas na seção "Regras de negócio" do README e cobertas por
`TestValorFinalDevolucaoAntecipadaNaoReduz`.

---

## Registro 5 — Auditoria do pipeline de CI

**Prompt**
> Verifique o que falta no pipeline de CI (`.github/workflows/ci.yml`) frente
> ao checklist da Atividade 3: instalação, teste real, build, pipeline verde,
> log de falha explicado, correção aplicada, bloqueio de merge em falha e
> registro de IA.

**Resposta (resumo)**
Apontou que o workflow instalava e testava mas não buildava nada; que nenhuma
branch tinha proteção configurada no GitHub (confirmado consultando a API,
não por suposição), então uma falha no CI não impedia merge; e que o README
ainda instruía `CI=true npm test -- --watchAll=false` (sintaxe do CRA), a
mesma flag que já havia quebrado o pipeline de verdade na migração para Vite
(execução com falha real, não simulada).

**Decisão da equipe**
Aceita. Adicionado o job `build` (compila a api e o bundle do frontend, só
após os testes passarem), corrigido README e Makefile para a sintaxe atual do
`vitest`, e configurada a exigência dos cinco status checks em `main` e `dev`
antes de permitir merge.

**Evidência de validação**
`go build ./cmd/app` e `npm run build` executados localmente com sucesso
antes do commit; pipeline verde e o incidente real de falha documentados em
[`validacao-e3.md`](validacao-e3.md).

---

## Registro 6 — Runner self-hosted, implantação e versionamento automático

**Prompt**
> Verifique se o runner self-hosted do repositório está funcionando e como
> demonstrá-lo. Depois: ao atualizar a `main`, o runner deve achar a versão
> mais recente e subir a nova versão; crie tags no git e versões automáticas
> que se reflitam no Docker Hub.

**Resposta (resumo)**
Consultou a API do GitHub e o serviço na máquina: o runner `jaime-note` estava
`online`, mas nunca tinha executado um job, porque todos os jobs da `main`
usavam `ubuntu-24.04` e o job de implantação do PR #28 só existia numa branch
com base na `homo`, em conflito com a `main`. Propôs um workflow manual de
demonstração (`runner-demo.yml`), depois levou o job de implantação para a
`main` e, por fim, o versionamento: `scripts/next-version.sh` calcula a versão
pelos Conventional Commits desde a última tag, a api é buildada com essa versão
(`-ldflags -X`), o CD publica `<serviço>-v<versão>`, cria a tag e a release no
GitHub e o runner implanta essa tag, conferindo o `/health`.

**Decisão da equipe**
Aceita, com correções no caminho:
- O disparo do `runner-demo.yml` falhou com `HTTP 404` na primeira tentativa: o
  arquivo ainda não estava na `main`, e `workflow_dispatch` só existe para
  workflows na branch padrão. Resolvido com PR para a `main` (#31).
- O job `release` da `homo` (PR #30), que lê a versão fixa do `server.go`,
  **não foi reaproveitado**: com versão manual, todo merge que esquecesse de
  mudar o número tentaria recriar a mesma tag. A versão passou a ser calculada.
- A versão é calculada antes do build, e não depois, para a imagem publicada
  continuar sendo a mesma que o smoke test validou (sem rebuild no CD).
- O próprio teste do script, escrito pela IA, entrou em recursão infinita
  (`commit()` chamando a si mesma após uma substituição de texto) e só foi
  corrigido porque foi executado antes do commit.

**Evidência de validação**
Run do workflow `Runner demo` com `Runner name: 'jaime-note'` e o `docker ps`
da máquina; run do merge do PR #32 com os jobs 1 a 8 verdes e a stack
recriada com `api-<sha>`; `bash scripts/next-version_test.sh` com os 8 casos
em `ok` (feat, fix, `!`, `BREAKING CHANGE`, corpo de merge, re-run e ausência
de tag); build local da api com `DRIVEFLOW_VERSION=9.9.9` respondendo
`"versao":"9.9.9"` no `/health`, e sem a variável mantendo `0.1.0`.

---

## O que a equipe não delegou à IA

Conforme a validação humana exigida no diagnóstico do E1:

- **Toda migration foi lida e revisada por uma pessoa** antes de entrar. A
  alteração do Registro 2 é o exemplo de por quê: a sugestão passava nos testes
  que existiam naquele momento e ainda assim estava errada para um caso real de
  negócio.
- **A decisão sobre as regras de locação** (cobrança por diária, política de
  multa, tratamento de devolução antecipada) é da equipe, não de sugestão
  automática.
- **O escopo da base e a escolha das branches** foram decididos pela equipe a
  partir da Atividade 2.
- **Nenhum resultado foi registrado como evidência sem ter sido executado.** O
  que ainda não rodou está marcado como pendente em
  [`validacao-e2.md`](validacao-e2.md).
