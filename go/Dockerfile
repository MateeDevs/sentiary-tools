# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go/go.mod go/go.sum ./
RUN go mod download
COPY go/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sentiary-mcp ./cmd/sentiary-mcp

FROM alpine:3.22 AS runtime
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /out/sentiary-mcp /app/sentiary-mcp
ENV PORT=8080
EXPOSE 8080
USER app
ENTRYPOINT ["/app/sentiary-mcp"]
