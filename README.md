## Create OpenAPI and jsonschma for kedge

### Creating OpenAPI

This will create OpenAPI configuration for kedge

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
