#!/bin/bash

for i in types spec gorilla; do oapi-codegen --package generated --generate $i -o pkg/server/generated/$i.go pkg/server/openapi/sever.spec.yaml ; done
