# ---- Etapa de construcao -------------------------------------------------
FROM node:22-alpine AS construcao

WORKDIR /app

# npm ci instala exatamente o que esta no package-lock.json, o que torna o
# build do container reproduzivel entre maquinas.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Vazio de proposito: as chamadas ficam relativas e o nginx repassa para a api.
ARG REACT_APP_API_URL=""
ENV REACT_APP_API_URL=$REACT_APP_API_URL

RUN npm run build

# ---- Imagem final --------------------------------------------------------
FROM nginx:1.27-alpine

COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=construcao /app/build /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=5 \
    CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
