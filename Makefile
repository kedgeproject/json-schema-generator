default: bin

.PHONY: all
all: bin

.PHONY: bin
bin:
	go build -o kedge-schema main.go parsego.go

.PHONY: install
install: bin
	cp kedge-schema $(GOBIN)/

.PHONY: container-image
container-image: bin
	docker build -t surajd/kedgeschema -f ./scripts/Dockerfile .

.PHONY: generate-config
generate-config: container-image
	docker run -v `pwd`:/data:Z surajd/kedgeschema
