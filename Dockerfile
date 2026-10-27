# syntax=docker/dockerfile:1

FROM python:3.13-slim AS docs-build
WORKDIR /docs
COPY requirements-docs.txt ./
RUN pip install --no-cache-dir -r requirements-docs.txt
COPY zensical.toml ./
COPY docs ./docs
COPY overrides ./overrides
COPY examples ./examples
RUN zensical build --strict

FROM nginx:stable-alpine AS docs
RUN apk add --no-cache ca-certificates
COPY --from=docs-build /docs/site /usr/share/nginx/html
COPY docs-nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80

FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sentiary-mcp ./cmd/sentiary-mcp

FROM alpine:3.22 AS runtime
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=build /out/sentiary-mcp /app/sentiary-mcp

ENV PORT=8080
EXPOSE 8080

USER app
ENTRYPOINT ["/app/sentiary-mcp"]
