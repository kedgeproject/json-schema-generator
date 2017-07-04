## Create OpenAPI and jsonschma for kedge

### Creating OpenAPI

This will create OpenAPI configuration for [kedge](https://github.com/kedgeproject/kedge)

```bash
$ go run main.go
```

### Creating JSONSchema

Install [openapi2jsonschema](https://github.com/garethr/openapi2jsonschema)

```bash
$ pip install openapi2jsonschema
```

Download the openAPI schema of Kubernetes

```bash
$ curl -O https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json
```

Generate OpenAPI for kedge

```bash
$ go run main.go > output.json
```

Create the json schema

```bash
$ mkdir configs
$ openapi2jsonschema output.json -o configs/ --stand-alone
```

### Validating against schema

Install [jsonschema tool](https://github.com/Julian/jsonschema) locally

```bash
$ jsonschema -F "{error.message}" -i ./configs/db.json ./configs/appspec.json 
```

The file [appspec.json](./configs/appspec.json) has schema for validating kedge.
The above file [db.json](./configs/db.json) is taken from [kedge repo example](https://github.com/kedgeproject/kedge/blob/master/examples/envFrom/db.yaml).
