#!/bin/bash

set -x

curl -O https://raw.githubusercontent.com/kubernetes/kubernetes/$(curl https://raw.githubusercontent.com/surajssd/kedgeSchema/master/scripts/k8s-release)/api/openapi-spec/swagger.json
# TODO: replace the download link below with the link from the upstream one
curl -O https://raw.githubusercontent.com/surajssd/kedgeSchema/master/spec.go

kedgeSchemaGen > output.json

mkdir -p configs
openapi2jsonschema output.json -o configs/ --stand-alone

rm -rf swagger.json output.json
