# build stage
FROM golang:1.9.0 AS build-env
ADD . /go/src/github.com/kedgeproject/kedge-jsonschema
RUN cd /go/src/github.com/kedgeproject/kedge-jsonschema && go build -o kedge-jsonschema main.go parsego.go

# final stage
FROM garethr/openapi2jsonschema

RUN apk add --update curl && \
    rm -rf /var/cache/apk/*
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY --from=build-env /go/src/github.com/kedgeproject/kedge-jsonschema/kedge-jsonschema /usr/local/bin/
COPY ./scripts/entrypoint.sh /usr/local/bin/

WORKDIR /data

ENTRYPOINT []
CMD ["/bin/sh", "/usr/local/bin/entrypoint.sh"]
 

