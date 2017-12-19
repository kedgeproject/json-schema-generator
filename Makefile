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
	go build schemagen.go

.PHONY: install
install:
	go install schemagen.go

.PHONY: container-image
container-image:
	docker build -t kedge/kedgeschema -f ./scripts/Dockerfile .

.PHONY: generate-config
generate-config:
	docker run -v `pwd`:/data:Z kedge/kedgeschema

.PHONY: test-generate-config
test-generate-config:
	docker run kedge/kedgeschema

.PHONY: install-gotools
install-gotools:
	go get github.com/Masterminds/glide
	go get github.com/sgotti/glide-vc

.PHONY: update-vendor
update-vendor: install-gotools
	glide update --strip-vendor
	glide-vc --only-code --no-tests

