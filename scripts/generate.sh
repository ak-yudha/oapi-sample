#!/bin/bash
set -e

mkdir -p internal/gen

go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -package gen -generate types,server,spec -o internal/gen/api.gen.go api/openapi.yaml
