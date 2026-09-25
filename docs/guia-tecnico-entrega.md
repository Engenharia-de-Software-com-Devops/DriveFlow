# Guia Técnico de Entrega — Docker e Pipeline CI/CD do Projeto Integrador

> Transcrição em Markdown do PDF *Desenvolvimento de Software Integrado — DevOps —
> Guia Técnico de Entrega (Docker e Pipeline CI-CD)*, Turma 4 — Z251,
> PG2305-04-Z251, semestre 2026.2. Em caso de divergência, vale o PDF original.

Este documento é o padrão técnico obrigatório para a entrega do Projeto
Integrador: o que publicar no repositório, como estruturar Dockerfile, Docker
Compose e o pipeline de CI/CD, e quais boas práticas de containers são exigidas.
Ele complementa — não substitui — o Enunciado do Projeto Integrador. Em caso de
dúvida sobre pesos, prazos ou critérios de nota, o Enunciado é a fonte de
verdade.

## 1. O que a equipe precisa entregar

No fim do Encontro 6, cada equipe entrega um único link de repositório público.
Toda evidência mora nele — nada é avaliado fora do repositório. Este guia
detalha o padrão técnico esperado nas partes de containers e pipeline, que
alimentam 45% da nota final (25% Pipeline CI + 20% Containers/Compose) e também
aparecem nos critérios de Projeto funcional e Entrega e operação.

| Item | O que precisa existir no repositório |
| ---- | ------------------------------------ |
| Repositório | Público, com histórico de commits real da equipe (GitHub Flow: branches curtas + Pull Request para a `main`). |
| `README.md` | Documentação completa da aplicação — ver seção 2. É o ponto de entrada do professor ao avaliar o repositório, e de qualquer pessoa que for rodar o projeto. |
| `Dockerfile` | Constrói a imagem da aplicação seguindo as boas práticas da seção 3. |
| `docker-compose.yml` | Sobe a aplicação e suas dependências (banco, cache etc.) com um único comando — ver seção 4. |
| Pipeline CI/CD | Workflow versionado (ex.: GitHub Actions) com etapas de CI e de CD claramente separadas — ver seção 5. |
| Imagem publicada | Imagem Docker da aplicação publicada em um registry público — Docker Hub é o recomendado — ver seção 6. |

> **Importante:** rodar em nuvem não é exigido. A "entrega contínua" deste
> projeto termina quando a imagem chega pronta e publicada no registry — é a
> partir dela que o professor roda a aplicação na própria máquina, na avaliação
> do Encontro 6. Isso é combinado com o Encontro 5 (Entrega e operação), que
> trabalha release notes e rollback sobre esse mesmo fluxo.

### Onde isso entra na nota final

Este guia é o padrão técnico de apenas 2 dos 6 critérios oficiais do Plano de
Ensino — os pesos abaixo já são os pesos oficiais e não podem ser alterados:

| Peso | Critério | Coberto por este guia? |
| ---- | -------- | ---------------------- |
| 25% | Pipeline CI | Sim — seção 5 |
| 20% | Containers/Compose | Sim — seções 3, 4 e 7 |
| 25% | Projeto funcional | Não — ver Enunciado do Projeto Integrador |
| 15% | Entrega e operação | Não — release notes e rollback, ver Encontro 5 |
| 10% | Diagnóstico DevOps | Não — ver Encontro 1 |
| 5% | Apresentação final | Não — ver Encontro 6 |

45% (Pipeline CI 25% + Containers/Compose 20%) é a parte da nota que este guia
cobre. 55% (Projeto funcional 25% + Entrega e operação 15% + Diagnóstico DevOps
10% + Apresentação final 5%) vem dos demais critérios.

## 2. O README — o que precisa conter

O README é a documentação da aplicação. Ele deve permitir que alguém que nunca
viu o projeto clone o repositório e coloque tudo para rodar — local ou a partir
da imagem publicada — sem precisar perguntar nada à equipe. Estrutura mínima
esperada, nesta ordem:

1. **Visão geral** — nome do projeto, o problema que resolve, principais
   tecnologias (linguagem, framework, banco de dados).
2. **Arquitetura** — um diagrama simples (pode ser um bloco de texto/ASCII)
   mostrando os componentes (API, banco, outras dependências) e como se
   comunicam.
3. **Como rodar localmente (sem Docker)**, se aplicável — pré-requisitos,
   variáveis de ambiente, comandos de instalação e execução.
4. **Como rodar com Docker Compose** — o caminho principal de avaliação:
   comandos exatos, do clone ao "funciona", incluindo como preencher variáveis
   (arquivo `.env.example` versionado, nunca segredos reais).
5. **Como rodar a partir da imagem publicada** — comando de `docker pull` com o
   nome completo da imagem no Docker Hub e o comando para executá-la (ou o
   `docker compose` apontando para a imagem publicada).
6. **Como rodar os testes** — comando único, sem passos manuais extras.
7. **Pipeline CI/CD** — o que cada etapa faz e onde ver as execuções (ex.: aba
   Actions do GitHub).
8. **Variáveis de ambiente** — tabela com nome, para que serve e valor de
   exemplo (nunca o valor real de produção/segredo).
9. **Uso de IA** — registro do que foi pedido a uma IA, o que foi aceito e o
   que foi corrigido. Mantra da disciplina: toda saída de IA é hipótese até ser
   validada por teste, execução ou revisão humana.
10. **Troubleshooting** — 2 ou 3 problemas comuns que a equipe encontrou (ex.:
    porta ocupada, container não conecta no banco) e como resolver.

> **Dica prática:** um README bom é aquele que resiste ao teste do "computador
> limpo": peça para alguém de outra equipe clonar o repositório e seguir só o
> que está escrito. Se travar, o README (não a pessoa) está incompleto.

## 3. Dockerfile — regras obrigatórias e boas práticas

Estas regras valem para toda imagem publicada pela equipe.

| Regra | Por quê |
| ----- | ------- |
| Imagem base leve | Prefira variantes slim/alpine (ex.: `node:22-bookworm-slim`, `python:3.12-slim`, `eclipse-temurin:21-jre-alpine`). Menos pacotes = imagem menor, build/pull mais rápido e menos superfície de vulnerabilidade. |
| Nunca rodar como root | Crie ou use um usuário não privilegiado (`USER` no final do Dockerfile). Se o container for comprometido, um processo root tem muito mais poder de causar dano. |
| Versão da imagem base fixada | Use uma tag específica, nunca `latest`. `latest` muda sem aviso e quebra builds reprodutíveis. |
| `.dockerignore` | Exclua `node_modules`, `.git`, `.env`, arquivos de teste e afins. Evita vazar segredos para a imagem e mantém o contexto de build pequeno. |
| Cache de camadas | Copie primeiro os arquivos de dependências e instale antes de copiar o código-fonte. Mudar código não invalida a camada de dependências. |
| Multi-stage build quando fizer sentido | Compile/instale em um estágio e copie só o artefato final para a imagem de execução. Reduz o tamanho e remove ferramentas de build da produção. |
| Nenhum segredo na imagem | Senhas, tokens e strings de conexão nunca vão para o Dockerfile nem para o código copiado. Entram em tempo de execução via variáveis de ambiente (`env_file`, secrets do CI). |
| Expor só a porta necessária | Um `EXPOSE` claro documenta o contrato do container. |
| Um processo principal por container | Cada serviço (API, banco, cache) é um container separado, orquestrado pelo Compose. |
| `HEALTHCHECK` recomendado | Ajuda o Compose (e o professor) a saber quando o serviço está de fato pronto, não só "o processo iniciou". |

Exemplo de referência (Node.js):

```dockerfile
FROM node:22-bookworm-slim
WORKDIR /app

# Copia só os manifestos primeiro -> aproveita cache de camadas
COPY package.json package-lock.json ./
RUN npm ci --omit=dev

# Só agora copia o código, já com dono no usuário não-root
COPY --chown=node:node . .

ENV NODE_ENV=production
USER node
EXPOSE 3000

HEALTHCHECK --interval=10s --timeout=3s CMD node -e "process.exit(0)"
CMD ["node", "server.js"]
```

Outras linguagens seguem a mesma lógica (imagem slim, usuário não-root,
dependências antes do código) — adapte os comandos, não os princípios.

## 4. Docker Compose — regras

O `docker-compose.yml` garante que a aplicação sobe do zero, de forma
repetível, sem passo manual extra: `docker compose up --build` precisa subir
tudo sozinho.

- **Um comando sobe tudo** — `docker compose up --build` não pode exigir nenhum
  passo manual antes (criar pasta, rodar script, editar arquivo). O que precisar
  ser preenchido vem de variável de ambiente com valor padrão de exemplo.
- **Serviços se falam pelo nome do serviço, nunca por `localhost`** — a API
  acessa o banco por `db:5432`. `localhost` dentro de um container aponta para
  o próprio container.
- **Persistência com volumes nomeados** — dados de banco usam um named volume,
  para sobreviver a `docker compose down` (sem `-v`) e a recriações.
- **Variáveis via `.env`** — valores versionados em `.env.example` (sem segredo
  real); o `.env` real fica fora do controle de versão (`.gitignore`).
- **`healthcheck` e `depends_on` com `condition`** — a API deve esperar o banco
  estar pronto (`condition: service_healthy`), não só "container iniciado".
- **Portas publicadas com intenção** — só publique (`ports:`) o que precisa ser
  acessado de fora do Compose; serviços internos (ex.: banco) podem ficar só na
  rede interna.

Exemplo de referência:

```yaml
services:
  api:
    build: .
    ports:
      - "3000:3000"
    environment:
      DATABASE_URL: postgres://app:app@db:5432/app
    depends_on:
      db:
        condition: service_healthy

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: app
      POSTGRES_PASSWORD: app
      POSTGRES_DB: app
    volumes:
      - db_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U app"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  db_data:
```

## 5. Pipeline CI/CD — estrutura obrigatória

O pipeline vive em um único workflow versionado (ex.:
`.github/workflows/ci-cd.yml`) com duas etapas nitidamente separadas: CI e CD.
CD só deve rodar depois que CI passar (`needs: ci` no GitHub Actions), porque
publicar uma imagem que falhou nos testes é pior do que não publicar nada.

| Etapa | O que faz | Quando roda |
| ----- | --------- | ----------- |
| CI (Integração Contínua) | Instala dependências, roda lint/testes automatizados, builda a imagem Docker (`docker build`, sem publicar) e valida a imagem subindo-a com o Compose para uma checagem real (ex.: a API responde em `/health`). | A cada push e a cada Pull Request. |
| CD (Entrega Contínua) | Autentica no registry, coloca tag na imagem já validada pelo CI e publica (`docker push`) no Docker Hub. Gera também as instruções/artefato de execução local (ver seção 7). | Só após CI passar, e só na branch principal (`main`), depois do merge do PR. |

> **Regra de ouro:** a imagem publicada no CD deve ser a mesma que passou no CI
> — não uma reconstruída do zero. Construa a imagem uma vez no CI e
> reaproveite-a no CD (`docker save`/`docker load`, ou push condicionado ao
> sucesso do job de CI).

### Boas práticas do pipeline

- **Segredos nunca em texto plano** — usuário e token do Docker Hub ficam em
  GitHub Secrets (ex.: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`), nunca no YAML.
- **Falha real interrompe o pipeline** — nenhum step usa `continue-on-error` ou
  `|| true` para "maquiar" uma falha; se algo quebrou, o pipeline fica vermelho.
- **Tag da imagem rastreável** — nunca publique só `latest`. Use ao menos o SHA
  do commit (ex.: `usuario/app:${{ github.sha }}`) e, opcionalmente, também
  `latest` e/ou uma versão semântica.
- **GitHub Flow** — branches curtas a partir da `main`, Pull Request
  obrigatório antes de mesclar; é o disparo natural de CI (no PR) e de CD (após
  o merge).

Exemplo comentado do encadeamento CI → CD no repositório da disciplina:
`Práticas/encontro-5-entrega/workflow-cd-exemplo.yml`.

## 6. Publicando no Docker Hub — passo a passo

Docker Hub é o registry recomendado (gratuito para imagens públicas, integra
direto com GitHub Actions). Outro registry público equivalente (ex.: GitHub
Container Registry) é aceito, desde que a imagem seja pública e o fluxo seja
análogo.

1. Criar uma conta gratuita em hub.docker.com (de um integrante ou da equipe).
2. Criar um repositório de imagem público (ex.: `seu-usuario/nome-do-projeto`).
3. Gerar um Access Token em *Account Settings → Security* (nunca usar a senha
   da conta no pipeline).
4. Cadastrar o usuário e o token como Secrets do repositório no GitHub
   (*Settings → Secrets and variables → Actions*).
5. No job de CD, autenticar, marcar a tag e publicar a imagem.

```yaml
- name: Login no Docker Hub
  uses: docker/login-action@v3
  with:
    username: ${{ secrets.DOCKERHUB_USERNAME }}
    password: ${{ secrets.DOCKERHUB_TOKEN }}

- name: Tag e push da imagem
  run: |
    docker tag app:${{ github.sha }} seu-usuario/nome-do-projeto:${{ github.sha }}
    docker tag app:${{ github.sha }} seu-usuario/nome-do-projeto:latest
    docker push seu-usuario/nome-do-projeto:${{ github.sha }}
    docker push seu-usuario/nome-do-projeto:latest
```

> **Atenção:** conferir, antes da entrega final, que a imagem realmente está
> pública no Docker Hub (`docker pull` sem estar logado, de outra máquina, é o
> teste mais simples). Uma imagem privada que o professor não consegue baixar
> equivale a não ter entregado.

## 7. O artefato de entrega — rodando localmente

Este projeto não exige hospedar a aplicação na internet. O CD termina quando a
imagem chega publicada e pronta — é a partir dela que o professor roda a
aplicação na própria máquina, no Encontro 6.

### Caminho principal: puxar do Docker Hub

O README (seção 2, item 5) deve documentar exatamente estes dois comandos, com
o nome real da imagem da equipe:

```bash
docker pull seu-usuario/nome-do-projeto:latest
docker compose -f docker-compose.prod.yml up
```

Quando a aplicação tem dependências (banco, cache), forneça um segundo arquivo
de Compose (ex.: `docker-compose.prod.yml`) que aponta para a imagem publicada
em vez de buildar localmente — troca-se `build: .` por
`image: seu-usuario/nome-do-projeto:latest`.

### Caminho alternativo (opcional): artefato do próprio pipeline

O CD também pode exportar a imagem já validada como artefato do GitHub Actions
(`docker save` + `actions/upload-artifact`), para o professor baixar e carregar
sem depender do Docker Hub. É um complemento, não substitui o registry.

```bash
docker save --output imagem.tar seu-usuario/nome-do-projeto:latest
# ...
docker load --input imagem.tar
```

## 8. Checklist final antes de entregar

- [ ] Repositório público, com README completo (seção 2).
- [ ] Dockerfile usa imagem base slim/alpine, versão fixada, roda com usuário
      não-root e não copia segredos.
- [ ] `docker compose up --build` sobe a aplicação inteira sem nenhum passo
      manual extra.
- [ ] Pipeline tem CI e CD como etapas separadas, e CD só roda depois que CI
      passa.
- [ ] Segredos do Docker Hub estão em GitHub Secrets, não no YAML.
- [ ] Imagem está publicada e pública no Docker Hub, com tag rastreável (sha
      e/ou versão).
- [ ] README documenta o comando exato para puxar e rodar a imagem publicada,
      localmente.

> **Regra de ouro:** o que não estiver pronto e verificável no repositório no
> fim do Encontro 6 não conta — a régua é sempre "o entregável existe e
> funciona?" antes de "está bem documentado?".
