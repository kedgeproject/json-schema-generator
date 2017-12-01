# build stage
FROM golang:1.9.0 AS build-env
ADD . /go/src/github.com/kedgeproject/json-schema-generator
RUN export GOBIN=$GOPATH/bin && cd /go/src/github.com/kedgeproject/json-schema-generator && make install

# final stage
FROM surajd/openapi2jsonschema:generate-configs-for-kedge
COPY --from=build-env /go/bin/schemagen /usr/local/sbin/

RUN dnf install -y --setopt=tsflags=nodocs npm && \
	dnf clean all && \
	npm install -g api-spec-converter --unsafe-perm=true --allow-root 

COPY ./scripts/entrypoint.sh /usr/local/sbin/

WORKDIR /data

CMD ["entrypoint.sh"]
 
