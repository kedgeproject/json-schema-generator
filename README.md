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
jsonschema -F "{error.message}" -i ./example/db.json ./configs/deploymentspecmod.json
```
Since the input file [db.json](./example/db.json) has deployment so we are using this
[deploymentspecmod.json](./configs/deploymentspecmod.json) to validate. If the controller
is different we will have to use a different file.


The file [deploymentspecmod.json](./configs/deploymentspecmod.json) has schema for validating kedge.
The above file [db.json](./example/db.json) is taken from [kedge repo example](https://github.com/kedgeproject/kedge/blob/master/examples/envFrom/db.yaml).
