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

CONTAINERIMAGE=kedge/json-schema-generator:latest

default: bin

.PHONY: all
all: bin

.PHONY: bin
bin:
	go build schemagen.go

.PHONY: install
install:
	go install schemagen.go

.PHONY: container-image
container-image:
	docker build -t ${CONTAINERIMAGE} -f ./scripts/Dockerfile .

.PHONY: generate-config
generate-config:
	mkdir -p _output
	cd _output && docker run --rm -v `pwd`:/data:Z ${CONTAINERIMAGE}

.PHONY: test-generate-config
test-generate-config:
	docker run ${CONTAINERIMAGE}

.PHONY: install-gotools
install-gotools:
	go get github.com/Masterminds/glide
	go get github.com/sgotti/glide-vc

.PHONY: update-vendor
update-vendor: install-gotools
	glide update --strip-vendor
	glide-vc --only-code --no-tests

.PHONY: generate-config-local
generate-config-local: install
	mkdir -p _output
	cd _output && ../scripts/entrypoint.sh

.PHONY: generate-config-local-strict
generate-config-local-strict: install
	mkdir -p _output
	cd _output && ../scripts/entrypoint.sh --strict

.PHONY: clean-output
clean-output:
	rm -rf _output
