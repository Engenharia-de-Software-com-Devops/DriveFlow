# ---- Etapa de construcao -------------------------------------------------
FROM node:22-alpine AS build

WORKDIR /app

# As dependencias entram antes do codigo para aproveitar o cache de camadas.
COPY package*.json ./
RUN npm ci

COPY . .

# Vazio por padrao: o bundle chama /api na mesma origem e o nginx repassa.
ARG VITE_API_URL=""
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# ---- Imagem final --------------------------------------------------------
# Variante do nginx que roda sem root e escuta na 8080 (porta nao privilegiada).
FROM nginxinc/nginx-unprivileged:1.27-alpine

COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

USER nginx
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
    CMD wget --quiet --tries=1 --spider http://127.0.0.1:8080/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
