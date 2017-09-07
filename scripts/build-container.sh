#!/bin/bash

go build -o kedgeSchemaGen main.go parsego.go

docker build -t surajd/kedgeschemagen -f ./scripts/Dockerfile .
docker run -v `pwd`:/data:Z surajd/kedgeschemagen




