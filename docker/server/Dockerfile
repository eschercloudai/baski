# syntax=docker/dockerfile:1
FROM cgr.dev/chainguard/go:latest as go-builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd/
COPY pkg ./pkg/

RUN go mod download

RUN CGO_ENABLED=0 go build -o /tmp/baski-server /app/cmd/server/main.go


FROM cgr.dev/chainguard/wolfi-base:latest

ENV S3_ENDPOINT=""
ENV S3_ACCESSKEY=""
ENV S3_SECRETKEY=""
ENV S3_BUCKET="baski"

RUN apk update --no-cache
RUN apk add jq curl

RUN echo "baski:x:1000:1000:Baski Non Root,,,:/home/baski:" >> /etc/passwd
RUN mkdir -p /home/baski/
RUN chown baski: -R /home/baski

COPY --from=go-builder /tmp/baski-server /usr/bin/

RUN chmod 755 /usr/bin/baski-server
ENV PATH=${PATH}:/home/baski/.local/bin/

USER 1000

ENTRYPOINT ["baski-server"]
