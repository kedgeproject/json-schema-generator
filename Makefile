default: bin

.PHONY: all
all: bin

.PHONY: bin
bin:
	go build -o kedgeSchemaGen main.go parsego.go

.PHONY: install
install: bin
	cp kedgeSchemaGen $(GOBIN)/

.PHONY: container-image
container-image: bin
	docker build -t surajd/kedgeschemagen -f ./scripts/Dockerfile .

.PHONY: generate-config
generate-config: container-image
	docker run -v `pwd`:/data:Z surajd/kedgeschemagen
