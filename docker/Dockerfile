FROM alpine:3.17.1

RUN apk update --no-cache
RUN apk add bash

COPY bin/amd64-linux-gnu/baski /
RUN echo "baski:x:1000:1000:Baski Non Root,,,::" >> /etc/passwd
USER 1000

ENTRYPOINT ["bash"]