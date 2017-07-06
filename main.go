package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/go-openapi/spec"
	"k8s.io/apimachinery/pkg/openapi"
)

var kedgeSpec = `{
  "io.kedge.AppSpec": {
    "description": "AppSpec is a description of a app.",
    "required": [
      "containers",
      "name"
    ],
    "properties": {
      "name": {
        "description": "Name of the micro-service",
        "type": "string"
      },
      "labels": {
        "description": "Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels",
        "type": "object",
        "additionalProperties": {
         "type": "string"
        }
      },
      "persistentVolumes": {
        "description": "List of persistentVolumes that should be mounted on the pod.",
        "type": "array",
        "items": {
          "$ref": "#/definitions/io.kedge.PersistentVolume"
        }
      },
      "configMaps": {
        "description": "List of configMaps.",
        "type": "array",
        "items": {
          "$ref": "#/definitions/io.kedge.ConfigMap"
        }
      },
      "services": {
        "description": "List of Kubernetes Services.",
        "type": "array",
        "items": {
          "$ref": "#/definitions/io.kedge.ServiceSpec"
        }
      },
      "ingresses": {
        "description": "List of Kubernetes Ingress.",
        "type": "array",
        "items": {
          "$ref": "#/definitions/io.kedge.IngressSpec"
        }
      },
      "containers": {
        "description": "List of containers belonging to the pod. Containers cannot currently be added or removed. There must be at least one container in a Pod. Cannot be updated.",
        "type": "array",
        "items": {
          "$ref": "#/definitions/io.kedge.ContainerSpec"
        }
      }
    }
  },
  "io.kedge.PersistentVolume": {
    "description": "Define Persistent Volume to use in the app",
    "required": [
      "size"
    ],
    "properties": {
      "name": {
        "description": "Name of the persistent volume",
        "type": "string"
      },
      "size": {
        "description": "Size of persistent volume",
        "type": "string"
      }
    }
  },
  "io.kedge.ConfigMap": {
    "description": "Define ConfigMap to be created",
    "required": [
      "data"
    ],
    "properties": {
      "name": {
        "description": "Name of the configMap",
        "type": "string"
      },
     "data": {
      "description": "Data contains the configuration data. Each key must consist of alphanumeric characters, '-', '_' or '.'.",
      "type": "object",
      "additionalProperties": {
       "type": "string"
      }
     }
    }
  },
  "io.kedge.ServiceSpec": {
    "description": "Define Kubernetes service",
    "required": [
      "ports"
    ],
    "properties": {
      "name": {
        "description": "Name of the service",
        "type": "string"
      },
     "ports": {
      "description": "The list of ports that are exposed by this service. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies",
      "type": "array",
      "items": {
       "$ref": "#/definitions/io.kedge.ServicePort"
      }
     }
    }
  },
  "io.kedge.ServicePort": {
    "description": "Define service port",
    "required": [
      "port"
    ],
    "properties": {
      "endpoint": {
        "description": "Host to create ingress automatically",
        "type": "string"
      }
    }
  },
  "io.kedge.IngressSpec": {
    "description": "Create ingress object",
    "properties": {
      "name": {
        "description": "Name of the ingress",
        "type": "string"
      }
    }
  },
  "io.kedge.ContainerSpec": {
    "description": "A single application container that you want to run within a pod.",
    "required": [
     "name",
     "image"
    ],
    "properties": {
      "health": {
        "description": "Periodic probe of container liveness and readiness. Container will be restarted if the probe fails. Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes",
        "$ref": "#/definitions/io.k8s.api.core.v1.Probe"
      }
    }
  }
}`

func main() {
	filename := "swagger.json"

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("cannot read file %q: %v\n", filename, err)
	}

	api := &openapi.OpenAPIDefinition{}
	err = json.Unmarshal(content, &api.Schema)

	defs := generateBarekedgeSpec(api.Schema.SchemaProps.Definitions)

	// add defs to openapi
	for k, v := range defs {
		api.Schema.SchemaProps.Definitions[k] = v
	}
	PrintDefs(api.Schema)
}

func PrintDefs(v interface{}) {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(string(b))
}

func augmentProperties(s, t spec.Schema) {
	for k, v := range s.Properties {
		if _, ok := t.Properties[k]; !ok {
			t.Properties[k] = v
		}
	}
}

func generateBarekedgeSpec(k8sSpec spec.Definitions) spec.Definitions {
	// read the string into internal representation
	defs := spec.Definitions(make(map[string]spec.Schema))
	if err := json.Unmarshal([]byte(kedgeSpec), &defs); err != nil {
		fmt.Println(err)
	}

	// In `io.kedge.ServicePort` add `io.k8s.api.core.v1.ServicePort`
	// In `io.kedge.AppSpec` add `io.k8s.api.core.v1.PodSpec` and `io.k8s.api.extensions.v1beta1.DeploymentSpec`
	// In `io.kedge.PersistentVolume` add `io.k8s.api.core.v1.PersistentVolumeClaimSpec`
	// In `io.kedge.ServiceSpec` add `io.k8s.api.core.v1.ServiceSpec`
	// In `io.kedge.ServicePort` add `io.k8s.api.core.v1.ServicePort`
	// In `io.kedge.IngressSpec` add `io.k8s.api.extensions.v1beta1.IngressSpec`
	// In `io.kedge.ContainerSpec` add `io.k8s.api.core.v1.Container
	augmentMapping := []struct {
		t string
		s string
	}{
		{"io.kedge.AppSpec", "io.k8s.api.core.v1.PodSpec"},
		{"io.kedge.AppSpec", "io.k8s.api.extensions.v1beta1.DeploymentSpec"},
		{"io.kedge.PersistentVolume", "io.k8s.api.core.v1.PersistentVolumeClaimSpec"},
		{"io.kedge.ServiceSpec", "io.k8s.api.core.v1.ServiceSpec"},
		{"io.kedge.ServicePort", "io.k8s.api.core.v1.ServicePort"},
		{"io.kedge.IngressSpec", "io.k8s.api.extensions.v1beta1.IngressSpec"},
		{"io.kedge.ContainerSpec", "io.k8s.api.core.v1.Container"},
	}
	// then using th custom logic start adding things
	for _, m := range augmentMapping {
		augmentProperties(k8sSpec[m.s], defs[m.t])
	}
	return defs
}
