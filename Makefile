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
generate-config: container-image
	docker run -v `pwd`:/data:Z surajd/kedgeschema
