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
OS_OPENAPI_URL=https://raw.githubusercontent.com/openshift/origin/1252cce6daeca1b6cc0fd90b1bde5dcdc9a0853b/api/swagger-spec/oapi-v1.json
KEDGE_SPEC_URL=https://raw.githubusercontent.com/kedgeproject/kedge/master/pkg/spec/types.go
K8S_OPENAPI_FILE=k8s-oapi.json
OS_OPENAPI_V1_FILE=os-oapiv1.json
OS_OPENAPI_FILE=os-oapi.json
KEDGE_SPEC_FILE=kedge-types.go
KEDGE_OPENAPI_FILE=kedge-oapi.json
STRICT=false

echo "Downloading OpenAPI schema of Kubernetes from: $K8S_OPENAPI_URL"
curl -o $K8S_OPENAPI_FILE -z $K8S_OPENAPI_FILE $K8S_OPENAPI_URL
echo "Downloading Swagger schema of OpenShift from: $OS_OPENAPI_URL"
curl -o $OS_OPENAPI_V1_FILE -z $OS_OPENAPI_V1_FILE $OS_OPENAPI_URL
echo "Download Kedge types from: $KEDGE_SPEC_URL"
curl -o $KEDGE_SPEC_FILE -z $KEDGE_SPEC_FILE $KEDGE_SPEC_URL

echo "Converting Swagger schema for OpenShift to OpenAPI"
api-spec-converter $OS_OPENAPI_V1_FILE --from=swagger_1 --to=swagger_2 > $OS_OPENAPI_FILE
exit_status=$?
if [ $exit_status -ne 0 ]; then
	echo "Swagger to OpenAPI conversion failed"
	exit $exit_status
fi

echo "Generating OpenAPI schema for Kedge"
schemagen --kedgespec $KEDGE_SPEC_FILE --k8sSchema $K8S_OPENAPI_FILE --osSchema $OS_OPENAPI_FILE > $KEDGE_OPENAPI_FILE
exit_status=$?
if [ $exit_status -ne 0 ]; then
	echo "OpenAPI schema generation for Kedge failed"
	exit $exit_status
fi

echo "Generating JSONSchema for Kedge"
if [ "$1" == "--strict" ]; then
	echo "Setting strict mode for JSON Schema conversion"
	STRICT=true
fi
mkdir -p schema
if [ "$STRICT" = true ]; then
	openapi2jsonschema $KEDGE_OPENAPI_FILE -o schema/ --stand-alone --strict --kubernetes
else
	openapi2jsonschema $KEDGE_OPENAPI_FILE -o schema/ --stand-alone --kubernetes
fi
exit_status=$?
if [ $exit_status -ne 0 ]; then
	echo "Kedge JSONSchema generation failed"
	exit $exit_status
fi

echo "Copying controller JSON Schema files to controllers/ directory"
mkdir -p controllers
cp -rv schema/{deploymentconfigspecmod.json,jobspecmod.json,deploymentspecmod.json} controllers/

echo "Kedge JSONSchema generated successfully"
