#!/bin/bash

# Copyright 2017 The Kedge Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

K8S_OPENAPI_URL=https://raw.githubusercontent.com/kubernetes/kubernetes/$(curl https://raw.githubusercontent.com/kedgeproject/json-schema-generator/master/scripts/k8s-release)/api/openapi-spec/swagger.json
KEDGE_SPEC_URL=https://raw.githubusercontent.com/kedgeproject/kedge/master/pkg/spec/spec.go

echo "Downloading OpenAPI schema of Kubernetes from: $K8S_OPENAPI_URL"
curl -O $K8S_OPENAPI_URL

# Test if the spec.go exists, if it does don't download the file from URL
cat spec.go > /dev/null
if [ $? -ne 0 ]; then
	echo "Downloading 'spec.go' from $KEDGE_SPEC_URL"
	curl -O $KEDGE_SPEC_URL
else
	echo "'spec.go' already exists."
fi

echo "Generating OpenAPI schema for Kedge"
kedge-jsonschema > output.json

echo "Generating JSONSchema for Kedge"
mkdir -p schema
openapi2jsonschema --strict output.json -o schema/ --stand-alone

echo "Kedge JSONSchema generated successfully"
