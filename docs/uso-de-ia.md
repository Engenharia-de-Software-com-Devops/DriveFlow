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
