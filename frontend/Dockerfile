FROM node:22-bookworm AS build
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM nginx:1.27-alpine
LABEL org.opencontainers.image.source="https://github.com/Irico17/Areas-verdes-pucp"
COPY frontend/nginx.http.conf frontend/nginx.tls.conf frontend/nginx-headers.conf /opt/campus/
COPY frontend/docker-entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh && rm -f /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist /usr/share/nginx/html
EXPOSE 80 443
ENTRYPOINT ["/entrypoint.sh"]
