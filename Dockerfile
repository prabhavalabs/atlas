# syntax=docker/dockerfile:1.9

FROM node:24.7.0-alpine3.22 AS web-build
WORKDIR /src/web
RUN corepack enable && corepack prepare pnpm@11.11.0 --activate
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store pnpm install --frozen-lockfile
COPY web/ ./
ARG VITE_API_BASE_URL=https://atlas-api.prabhavalabs.com
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
RUN pnpm build

FROM golang:1.26.5-alpine AS go-build
WORKDIR /src
RUN apk add --no-cache ca-certificates tzdata
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
ARG ATLAS_VERSION=dev
ARG ATLAS_COMMIT=unknown
ARG ATLAS_BUILD_TIME=unknown
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath \
      -ldflags="-s -w -X github.com/prabhavalabs/atlas/internal/platform/buildinfo.version=${ATLAS_VERSION} -X github.com/prabhavalabs/atlas/internal/platform/buildinfo.commit=${ATLAS_COMMIT} -X github.com/prabhavalabs/atlas/internal/platform/buildinfo.buildTime=${ATLAS_BUILD_TIME}" \
      -o /out/atlas ./cmd/atlas

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=go-build /out/atlas /usr/local/bin/atlas
COPY --from=web-build /src/web/dist /app/web
USER nonroot:nonroot
EXPOSE 8080
ENV ATLAS_WEB_ROOT=/app/web \
    ATLAS_HTTP_ADDRESS=0.0.0.0:8080
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=4 \
  CMD ["/usr/local/bin/atlas", "healthcheck"]
ENTRYPOINT ["/usr/local/bin/atlas"]
CMD ["serve"]
