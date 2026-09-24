FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
ENV CGO_ENABLED=0
RUN go build -o /out/api ./cmd/api \
 && go build -o /out/migrate ./cmd/migrate \
 && go build -o /out/etl ./cmd/etl

FROM debian:bookworm-slim
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/api /out/migrate /out/etl /usr/local/bin/
COPY apps/api/migrations /opt/campus/migrations
COPY apps/api/openapi.yaml /opt/campus/openapi.yaml
COPY apps/api/docker-entrypoint.sh /entrypoint.sh
COPY data/raw /opt/campus/raw
COPY data/osm /opt/campus/osm
COPY data/mocks /opt/campus/mocks
RUN chmod +x /entrypoint.sh
ENV MIGRATIONS_DIR=/opt/campus/migrations \
    OPENAPI_PATH=/opt/campus/openapi.yaml \
    DATA_RAW_DIR=/opt/campus/raw \
    DATA_V1_DIR=/data/v1 \
    EDIFICIOS_PATH=/opt/campus/osm/edificios_pando.geojson \
    RESERVAS_MOCK_PATH=/opt/campus/mocks/reservas_agenda.mock.json \
    EVIDENCIAS_DIR=/data/evidencias \
    API_ADDR=:8091
EXPOSE 8091
ENTRYPOINT ["/entrypoint.sh"]
