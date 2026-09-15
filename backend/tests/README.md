# Testes do backend

Os testes ficam fora de `internal/`, espelhando as camadas da arquitetura, e
separados pelo único critério que muda a forma de rodar: **precisa de banco no
ar ou não.**

```
tests/
├── apoio/                    montagem compartilhada (não é arquivo de teste)
├── unidade/                  roda sem banco e sem Docker
│   ├── entities/             regras de tarifa e sobreposição de período
│   ├── usecases/             regras de negócio sobre o repositório em memória
│   ├── repository/           migrations lidas do disco, sem conexão
│   └── delivery/             rotas HTTP sobre o repositório em memória
└── integracao/               exige PostgreSQL no ar
    └── repository/           repositório PostgreSQL e migrations aplicadas
```

## Como rodar

```bash
make testar-backend      # só unidade: rápido, sem Docker
make testar-integracao   # só integração: sobe o container db
```

## Onde colocar um teste novo

| A situação | O lugar |
| ---------- | ------- |
| Cálculo ou invariante de domínio | `unidade/entities/` |
| Regra de negócio, validação, fluxo do contrato | `unidade/usecases/` |
| Status HTTP, formato do corpo, tradução de erro | `unidade/delivery/` |
| Conteúdo ou ordem das migrations, sem conectar | `unidade/repository/` |
| SQL de verdade, constraint do banco, migration aplicada | `integracao/repository/` |

Todo arquivo em `integracao/` começa com `//go:build integracao` na primeira
linha, antes do `package`. É isso que mantém o `go test ./...` do dia a dia sem
dependência de banco: sem a tag, o arquivo não entra na compilação.

## Duas restrições do Go que explicam o formato

1. **Teste em pasta separada só enxerga a API exportada** do pacote que testa.
   Foi possível mover tudo porque nenhum teste dependia de identificador não
   exportado. Se algum dia um teste precisar de interno, ele volta para a pasta
   do código como `_test.go` — a pasta do código continua sendo o lugar válido,
   e isso não é derrota nenhuma.
2. **Arquivo `_test.go` não é visível para outro pacote.** Por isso o apoio
   compartilhado entre `unidade/usecases` e `integracao/repository` mora em
   `apoio/servicos.go`, um arquivo normal, e não em um `_test.go`.

Como os testes não ficam mais ao lado do código, `go test` não atribui
cobertura automaticamente aos pacotes de `internal/`. Para medir cobertura, diga
quais pacotes contar:

```bash
cd backend && go test ./tests/unidade/... -coverpkg=./internal/... -cover
```
