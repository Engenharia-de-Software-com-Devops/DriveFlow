# DriveFlow

Projeto da disciplina de Engenharia de Software com DevOps.

## Estrutura

```
DriveFlow/
├── backend/    API em Go
└── frontend/   SPA em React (Create React App)
```

## Pre-requisitos

- [Go](https://go.dev/dl/) 1.25+
- [Node.js](https://nodejs.org/) 18+ e npm

## Como rodar

### Backend

```bash
cd backend
go run .
```

### Frontend

```bash
cd frontend
npm install
npm start
```

O frontend sobe em `http://localhost:3000`.

## Scripts do frontend

| Comando         | Descricao                                  |
| --------------- | ------------------------------------------ |
| `npm start`     | Ambiente de desenvolvimento com hot reload |
| `npm test`      | Executa os testes                           |
| `npm run build` | Gera o build de producao em `build/`        |
