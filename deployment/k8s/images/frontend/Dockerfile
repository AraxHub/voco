FROM node:22-alpine AS build

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci

COPY frontend/ ./

# Public (non-secret) Vite build args — override at docker build time.
ARG VITE_API_BASE_URL=
ARG VITE_KEYCLOAK_URL=https://auth.voco-online.ru
ARG VITE_KEYCLOAK_REALM=voco
ARG VITE_KEYCLOAK_CLIENT_ID=voco-frontend

ENV VITE_API_BASE_URL=$VITE_API_BASE_URL \
    VITE_KEYCLOAK_URL=$VITE_KEYCLOAK_URL \
    VITE_KEYCLOAK_REALM=$VITE_KEYCLOAK_REALM \
    VITE_KEYCLOAK_CLIENT_ID=$VITE_KEYCLOAK_CLIENT_ID

RUN npm run build

FROM nginx:1.29-alpine

COPY deployment/k8s/images/frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
