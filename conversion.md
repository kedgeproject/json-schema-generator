## Intro to JSONSchema

JSON Schema is a vocabulary that allows you to annotate and validate JSON/YAML
documents. 

- describes your existing data format
- clear, human- and machine-readable documentation
- complete structural validation, useful for
- automated testing
- validating client-submitted data

## Intro to OpenAPI or SwaggerSchame

The OpenAPI Specification (OAS) defines a standard, language-agnostic interface to
RESTful APIs which allows both humans and computers to discover and understand the
capabilities of the service without access to source code, documentation, or
through network traffic inspection. When properly defined, a consumer can
understand and interact with the remote service with a minimal amount of
implementation logic.

An OpenAPI definition can then be used by documentation generation tools to
display the API, code generation tools to generate servers and clients in various
programming languages, testing tools, and many other use cases.


## Where do we get the OpenAPI for kedge from?

This could be an obvious question that comes to your mind because Kedge is not any
web service where we have defined RESTful APIs. So even if Kedge is not a service,
but it has it's language specification defined in the form or [golang structs](https://github.com/kedgeproject/kedge/blob/master/pkg/spec/types.go),
which forms the basis of the language. Also since we are embedding the golang
structs from upstream Kubernetes, and Kubernetes has it's own OpenAPI schema defined
for each struct. So for Kedge this is good news, because now we have less work.

So the complete OpenAPI definition looks like the following, if you just look at the
root keys: ([Kubernetes 1.7 OpenAPI definition](https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.7/api/openapi-spec/swagger.json))

```json
{
  "swagger": "2.0",
  "info": {...},
  "paths": {...},
  "definitions": {...},
  "securityDefinitions": {...},
  "security": {...}
}
```

Now the part of interest for us is the root level field `"definitions": {...}`, it
has definition of all Kubernetes structs in OpenAPI format. If we dig deeper in
the `definitions` part, you will see something like this:

```json
{
  "swagger": "2.0",
  "definitions": {
    "io.k8s.kubernetes.pkg.api.v1.Pod": {...},
    "io.k8s.kubernetes.pkg.api.v1.PodAffinity": {...},
    "io.k8s.kubernetes.pkg.api.v1.PodAffinityTerm": {...},
    "io.k8s.kubernetes.pkg.api.v1.PodAntiAffinity": {...},
    ...
    "io.k8s.kubernetes.pkg.apis.apps.v1beta1.Deployment": {...},
    "io.k8s.kubernetes.pkg.apis.apps.v1beta1.DeploymentCondition": {...},
    "io.k8s.kubernetes.pkg.apis.apps.v1beta1.DeploymentList": {...},
    ...
  },
}
```

So all we need to do is inject the definition of Pod from here into our defintion.
But before we do that kinda injection first we need to know what to inject where.
To identify that we have introduced a convention in Kedge schema's golang struct
definition which looks something like the following. More information about the
conventions used in Kedge's golang struct definition can be found [here](https://github.com/kedgeproject/kedge/blob/3ade8541e2209148c21c636464e12aea75e35e3a/docs/development.md#specgo-conventions).

```go
// Container defines a single application container that you want to run within a pod.
// kedgeSpec: io.kedge.ContainerSpec
type Container struct {
	// One common definitions for 'livenessProbe' and 'readinessProbe'
	// this allows to have only one place to define both probes (if they are the same)
	// Periodic probe of container liveness and readiness. Container will be restarted
	// if the probe fails. Cannot be updated. More info:
	// https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// ref: io.k8s.kubernetes.pkg.api.v1.Probe
	// +optional
	Health *api_v1.Probe `json:"health,omitempty"`
	// k8s: io.k8s.kubernetes.pkg.api.v1.Container
	api_v1.Container `json:",inline"`
}
```
Above code can be found [here](https://github.com/kedgeproject/kedge/blob/ecc7df1b0bb46ff43d5f830e08c0d0ddafd8e710/pkg/spec/types.go#L84-L97).

Line `api_v1.Container `json:",inline"`` has a comment above it
`k8s: io.k8s.kubernetes.pkg.api.v1.Container` in this
`io.k8s.kubernetes.pkg.api.v1.Container` is a key of the Container definition in
Kubernetes's OpenAPI schema. This is how we know that for the Kedge's container
definition, the definition of upstream container comes from
`io.k8s.kubernetes.pkg.api.v1.Container` key in comment.

Fields those are embedded are put in with `k8s` as the way to tell that this
definition comes from Kubernetes.


Similarly ```Health *api_v1.Probe `json:"health,omitempty"` ``` has multiple
comments one says `+optional` which means that this field in this struct while
generation of definition is marked optional. And then there are bunch of comments
above `Health` field which explains what it does. These comments then become
description of the field in OpenAPI schema.

Above struct definition you can also see the comment
`kedgeSpec: io.kedge.ContainerSpec`. Here `io.kedge.ContainerSpec` is the key for
the Kedge's Container definition in the final output of the OpenAPI for Kedge.

With help of these conventions and parsing of go code and injecting upstream
Kubernetes OpenAPI schema into the Kedge's OpenAPI schema we generate final
OpenAPI schema which is superset of the Kubernetes OpenAPI schema.

## JSONSchema for Kedge

This the easiest part, we hand off this work to the tool called
[openapi2jsonschema](https://github.com/garethr/openapi2jsonschema). This then
reads all the keys from the definition part and creates JSONSchema for each key in
the definitions part of the OpenAPI Schema for kedge. All the JSONSchemas for the
Kedge can be found at [github.com/kedgeproject/json-schema](https://github.com/kedgeproject/json-schema/tree/master/schema).

Depending on what is the input at root level there are multiple schema files like
for Kubernetes Deployment with kedge we have [deploymentspecmod.json](https://github.com/kedgeproject/json-schema/blob/master/schema/deploymentspecmod.json), [jobspecmod.json](https://github.com/kedgeproject/json-schema/blob/master/schema/jobspecmod.json).

## Ref:

- JSON Schema http://json-schema.org/
- OpenAPI Specification https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.0.md
- Kubernetes master OpenAPI definition https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json
- openapi2jsonschema - https://github.com/garethr/openapi2jsonschema
- kedgeproject/json-schema - https://github.com/kedgeproject/json-schema
