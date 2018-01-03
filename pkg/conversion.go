/*
Copyright 2017 The Kedge Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/go-openapi/spec"
	"k8s.io/apimachinery/pkg/openapi"
)

func ParseOpenAPIDefinition(filename string) (*openapi.OpenAPIDefinition, error) {

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %q: %v\n", filename, err)
	}

	api := &openapi.OpenAPIDefinition{}
	err = json.Unmarshal(content, &api.Schema)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling in OpenAPI definition: %v", err)
	}
	return api, nil
}

func MergeDefinitions(target, src *openapi.OpenAPIDefinition) {
	for k, v := range src.Schema.SchemaProps.Definitions {
		target.Schema.SchemaProps.Definitions[k] = v
	}
}

func Conversion(KedgeSpecLocation, KubernetesSchema, OpenShiftSchema string, controllerOnly bool) error {
	defs, mapping, err := GenerateOpenAPIDefinitions(KedgeSpecLocation)
	if err != nil {
		return err
	}

	k8sApi, err := ParseOpenAPIDefinition(KubernetesSchema)
	if err != nil {
		return fmt.Errorf("kubernetes: %v", err)
	}

	osApi, err := ParseOpenAPIDefinition(OpenShiftSchema)
	if err != nil {
		return fmt.Errorf("openshift: %v", err)
	}

	MergeDefinitions(k8sApi, osApi)
	api := k8sApi

	defs = InjectKedgeSpec(api.Schema.SchemaProps.Definitions, defs, mapping)

	// add defs to openapi
	for k, v := range defs {
		api.Schema.SchemaProps.Definitions[k] = v
	}

	if controllerOnly {
		retainOnlyControllers(api.Schema.Definitions)
	}

	PrintJSONStdOut(api.Schema)
	return nil
}

func retainOnlyControllers(definitions spec.Definitions) {
	for k := range definitions {
		switch k {
		case "io.kedge.DeploymentSpecMod", "io.kedge.DeploymentConfigSpecMod", "io.kedge.JobSpecMod":
			continue
		default:
			delete(definitions, k)
		}
	}
}

func augmentProperties(s, t spec.Schema) spec.Schema {
	for k, v := range s.Properties {
		if _, ok := t.Properties[k]; !ok {
			t.Properties[k] = v
		}
	}
	t.Required = AddListUniqueItems(t.Required, s.Required)
	return t
}

func InjectKedgeSpec(koDefinitions spec.Definitions, kedgeDefinitions spec.Definitions, mappings []Injection) spec.Definitions {
	for _, m := range mappings {
		kedgeDefinitions[m.Target] = augmentProperties(koDefinitions[m.Source], kedgeDefinitions[m.Target])

		switch m.Target {
		// special case, where if the key is io.kedge.DeploymentSpec
		// ignore the required field called template
		case "io.kedge.DeploymentSpecMod",
			"io.kedge.DeploymentConfigSpecMod",
			"io.kedge.JobSpecMod":
			v := kedgeDefinitions[m.Target]
			var final []string
			for _, r := range v.Required {
				if r != "template" {
					final = append(final, r)
				}
			}
			v.Required = final
			kedgeDefinitions[m.Target] = v
		case "io.kedge.ContainerSpec":
			containerDef := kedgeDefinitions[m.Target]
			for i, r := range containerDef.Required {
				if r == "name" {
					containerDef.Required[i] = containerDef.Required[len(containerDef.Required)-1]
					containerDef.Required = containerDef.Required[:len(containerDef.Required)-1]
				}
			}
			kedgeDefinitions[m.Target] = containerDef
		}
	}
	return kedgeDefinitions
}

func PrintJSONStdOut(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(b))
}
