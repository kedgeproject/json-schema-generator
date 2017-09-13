#!/bin/bash

set -xe

curl -O https://raw.githubusercontent.com/kubernetes/kubernetes/$(curl https://raw.githubusercontent.com/surajssd/kedgeSchema/master/scripts/k8s-release)/api/openapi-spec/swagger.json
#curl -O https://raw.githubusercontent.com/kedgeproject/kedge/master/pkg/spec/spec.go

kedgeSchemaGen > output.json

mkdir -p configs
openapi2jsonschema output.json -o configs/ --stand-alone

