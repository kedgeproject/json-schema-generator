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

default: bin

.PHONY: all
all: bin

.PHONY: bin
bin:
	go build -o kedge-jsonschema main.go parsego.go

.PHONY: install
install: bin
	cp kedge-jsonschema $(GOBIN)/

.PHONY: container-image
container-image:
	docker build -t surajd/kedgeschema -f ./scripts/Dockerfile .

.PHONY: generate-config
generate-config:
	docker run -v `pwd`:/data:Z surajd/kedgeschema
