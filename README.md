# Create JSONSchema for Kedge

## Creating OpenAPI

This will create OpenAPI configuration for [kedge](https://github.com/kedgeproject/kedge),
but make sure you have Kubernetes Swagger OpenAPI schema and Kedge `spec.go` downloaded
locally. For detailed steps see [manual steps](https://github.com/surajssd/kedge-schema#doing-it-the-hard-way).

```bash
make install
kedge-schema
```

## Creating JSONSchema

### Doing it the easy way

**Note**: Needs [docker](https://docs.docker.com/engine/installation/) to be installed on
your machine locally.

```bash
make generate-config
```

The docker image used as base for creating this image has [`openapi2jsonschema`](https://github.com/garethr/openapi2jsonschema)
installed in it. Creating a docker image makes it easier to reduce the steps needed to do
things manually for various tools.

### Doing it the hard way

Let's download the Kubernetes OpenAPI schema

```bash
cd $GOPATH/src/github.com/surajssd/kedge-schema
curl -O https://raw.githubusercontent.com/kubernetes/kubernetes/$(curl https://raw.githubusercontent.com/surajssd/kedge-schema/master/scripts/k8s-release)/api/openapi-spec/swagger.json
```

Also we need to download the Kedge [`spec.go`](https://github.com/kedgeproject/kedge/blob/master/pkg/spec/spec.go)
file

```bash
curl -O https://raw.githubusercontent.com/kedgeproject/kedge/master/pkg/spec/spec.go
```

Let's build the binary that generates OpenAPI schema for Kedge

```bash
make install
```

Generate OpenAPI schema for Kedge and save it in `output.json`

```bash
kedge-schema > output.json
```

This is just half done, now install a tool called [`openapi2jsonschema`](https://github.com/garethr/openapi2jsonschema).
It will read the OpenAPI specification stored in `output.json` and generate JSON Specification
for Kedge.

Once installed run `openapi2jsonschema`

```bash
mkdir -p configs
openapi2jsonschema output.json -o configs/ --stand-alone
```

Now all the JSONSchemas are generated in `configs` directory. The one that is most important
to us is `deploymentspecmod.json`.

**Protip**: To avoid all these manual steps do it the easy way.

## Validating against schema

Install [jsonschema tool](https://github.com/Julian/jsonschema) locally

```bash
jsonschema -F "{error.message}" -i ./example/db.json ./configs/deploymentspecmod.json
```
Since the input file [`db.json`](./example/db.json) has deployment so we are using this
[`deploymentspecmod.json`](./configs/deploymentspecmod.json) to validate. If the controller
is different we will have to use a different file.


The file [`deploymentspecmod.json`](./configs/deploymentspecmod.json) has schema for
validating kedge.
The above file [`db.json`](./example/db.json) is taken from [kedge repo example](https://github.com/kedgeproject/kedge/blob/master/examples/envFrom/db.yaml).
