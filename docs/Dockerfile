# syntax=docker/dockerfile:1

FROM python:3.13-slim AS docs-build
WORKDIR /src
COPY docs/requirements.txt ./docs/requirements.txt
RUN pip install --no-cache-dir -r docs/requirements.txt
COPY docs/ ./docs/
COPY go/examples/ ./go/examples/
RUN zensical build --strict --config-file docs/zensical.toml

FROM nginx:stable-alpine AS docs
RUN apk add --no-cache ca-certificates
COPY --from=docs-build /src/docs/site /usr/share/nginx/html
COPY docs/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
