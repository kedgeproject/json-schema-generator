## Create OpenAPI and jsonschma for kedge

### Creating OpenAPI

This will create OpenAPI configuration for [kedge](https://github.com/kedgeproject/kedge)

```bash
make install
kedgeSchemaGen
```

### Creating JSONSchema

```bash
make generate-config
```

### Validating against schema

Install [jsonschema tool](https://github.com/Julian/jsonschema) locally

```bash
jsonschema -F "{error.message}" -i ./example/db.json ./configs/appspec.json 
```

The file [appspec.json](./configs/appspec.json) has schema for validating kedge.
The above file [db.json](./example/db.json) is taken from [kedge repo example](https://github.com/kedgeproject/kedge/blob/master/examples/envFrom/db.yaml).
