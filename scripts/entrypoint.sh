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

set -xe

curl -O https://raw.githubusercontent.com/kubernetes/kubernetes/$(curl https://raw.githubusercontent.com/kedgeproject/json-schema-generator/master/scripts/k8s-release)/api/openapi-spec/swagger.json
curl -O https://raw.githubusercontent.com/kedgeproject/kedge/master/pkg/spec/spec.go

kedge-jsonschema > output.json

mkdir -p schema
openapi2jsonschema --strict output.json -o schema/ --stand-alone

