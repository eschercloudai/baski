# syntax=docker/dockerfile:1
FROM golang as go-builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd/
COPY pkg ./pkg/

RUN go mod download

RUN CGO_ENABLED=0 go build -o /tmp/baski /app/cmd/baski/main.go


FROM alpine:3.17

RUN apk update --no-cache
RUN apk add bash build-base jq curl git python3 py3-pip openssh-client

RUN echo "ServerAliveInterval 120" >> /etc/ssh/ssh_config

RUN echo "baski:x:1000:1000:Baski Non Root,,,:/home/baski:" >> /etc/passwd
RUN mkdir -p /home/baski/.config/openstack/
RUN chown baski: -R /home/baski

COPY --from=go-builder /tmp/baski /usr/bin/

RUN chmod 755 /usr/bin/baski
ENV PATH=${PATH}:/home/baski/.local/bin/

USER 1000

ENTRYPOINT ["bash"]
