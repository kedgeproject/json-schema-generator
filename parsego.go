package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
)

type sd struct {
	t string
	s string
}

func ParseStruct(strct *ast.StructType, spc *ast.GenDecl, defs spec.Definitions, fset *token.FileSet) []sd {

	var mapping []sd
	key, desc := ParseStructComments(spc.Doc)

	if key == "" {
		return mapping
	}

	if _, ok := defs[key]; !ok {
		defs[key] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: desc,
				Properties:  make(map[string]spec.Schema),
			},
		}
	}

	// iterate all the fields of struct
	for _, sf := range strct.Fields.List {
		fmt.Println(sf.Names)
		ast.Fprint(os.Stdout, fset, sf, nil)

		//if sf.Names == nil {
		//	continue
		//}

		// get the field name from the json tag
		name, err := JSONTagName(sf.Tag.Value)
		if err != nil {
			panic(err)
		}
		fieldtype, err := GetStructFieldType(sf.Type)
		if err != nil {
			panic(err)
		}
		if fieldtype == "" {
			id, ok := sf.Type.(*ast.Ident)
			if !ok {
				continue
			}

			//obj, ok := id.Obj.Decl
			//if !ok {
			//	continue
			//}
			ts, ok := id.Obj.Decl.(*ast.TypeSpec)
			if !ok {
				continue
			}
			s, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			mapping = append(mapping, ParseStruct(s, spc, defs, fset)...)
			continue
		}

		// to extract the description use another function
		// also see if the field is optional
		desc, ref, optional := ParseStructFieldComments(sf.Doc)

		if fieldtype == "selectorExpr" {
			s := sd{t: key, s: ref}
			fmt.Println("adding s: ", s)
			mapping = append(mapping, s)
			continue
		}

		schema := spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: desc,
				Type:        spec.StringOrArray([]string{fieldtype}),
			},
		}

		if fieldtype == "object" {
			schema.AdditionalProperties = &spec.SchemaOrBool{
				Schema: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Type: spec.StringOrArray([]string{"string"}),
					},
				},
			}
		}
		if fieldtype == "array" {
			ref = "#/definitions/" + ref
			refObj, err := jsonreference.New(ref)
			if err != nil {
				panic(err)
			}
			schema.Items = &spec.SchemaOrArray{
				Schema: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Ref: spec.Ref{
							refObj,
						},
					},
				},
			}
		}

		if fieldtype == "starexpr" {
			ref = "#/definitions/" + ref
			refObj, err := jsonreference.New(ref)
			if err != nil {
				panic(err)
			}
			schema.Ref = spec.Ref{refObj}
			schema.Type = nil
		}

		defs[key].Properties[name] = schema
		if !optional && name != "" {
			f := defs[key]
			f.Required = append(f.Required, name)
			defs[key] = f
		}
	}
	return mapping
}

func main() {
	fset := token.NewFileSet() // positions are relative to fset
	defs := spec.Definitions(make(map[string]spec.Schema))

	// Parse the file containing this very example
	// but stop after processing the imports.
	node, err := parser.ParseFile(fset, "spec.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}
	var mapping []sd

	// search for the declaration of App struct
	for _, decl := range node.Decls {
		spc, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		//ast.Fprint(os.Stdout, fset, spc, nil)
		for _, s := range spc.Specs {
			stc, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}

			fmt.Printf("Struct Name: %s\ncomments: %s\n", stc.Name.Name, spc.Doc.Text())

			strct, ok := stc.Type.(*ast.StructType)
			if !ok {
				continue
			}
			mapping = append(mapping, ParseStruct(strct, spc, defs, fset)...)
		}
		PrintDefs(defs)
	}

	fmt.Println("mapping")
	for _, s := range mapping {
		fmt.Println(s.t, "-", s.s)
	}

	//PrintDefs(mapping)
	PrintDefs(defs)
	//ast.Fprint(os.Stdout, fset, node, nil)
}

// Returns string form of the struct field type
// if the type is not recognized then errors out
func GetStructFieldType(g interface{}) (string, error) {
	switch v := g.(type) {
	case *ast.Ident:
		// normal fields like following are identifiers
		// e.g.
		// Name string `json:"name"`
		// TODO: handle other types like int or bool
		if v.Name == "string" {
			return "string", nil
		} else {
			// this could cause problems for types that are identifiers and not string
			// so not adding it now because fields that are defined in same package
			// and embedded will cause problems e.g.
			// PodSpecMod `json:",inline"`
			// for builtin types keep adding checks in this if block
			// not adding check for 'PodSpecMod' so that we don't have to
			// edit this on every new addition of our defined field
			return "", nil
		}
	case *ast.MapType:
		// fields like following are of map type
		// e.g.
		// Data map[string]string `json:"data,omitempty"`
		key, ok1 := v.Key.(*ast.Ident)
		value, ok2 := v.Value.(*ast.Ident)
		if ok1 && ok2 {
			// TODO: only checking for string key and value
			// if needed also add other types of maps
			if key.Name == "string" && value.Name == "string" {
				return "object", nil
			} else {
				return "", fmt.Errorf("map types not string")
			}
		} else {
			// if maps either key or value is not identifier then
			// maybe handle it differently
			return "", fmt.Errorf("map key or value not identifier")
		}
	case *ast.ArrayType:
		// e.g.
		// Ports []ServicePortMod `json:"ports"`
		// above types are arrays
		return "array", nil
	case *ast.SelectorExpr:
		// if the type is from another package then it is of this type
		// e.g.
		// api_v1.PersistentVolumeClaimSpec `json:",inline"`
		return "selectorExpr", nil
	case *ast.StarExpr:
		// A StarExpr node represents an expression of the form "*" Expression.
		// Semantically it could be a unary "*" expression, or a pointer type.
		// e.g.
		// ConfigMapRef *ConfigMapEnvSource `json:"configMapRef,omitempty"`
		return "starexpr", nil
	default:
		// if none of above is satisfied then it should be added later
		// so keeping error so that we know that it is missing
		return "", fmt.Errorf("unknown type could not identify")
	}
}

// If given a JSON tag this will extract struct field name
// out of it. For e.g. if a json tag is like this
// `json:"persistentVolumes,omitempty"`
// This function will return 'persistentVolumes'
func JSONTagName(j string) (string, error) {
	// The tag that we get has double quotes which are escaped
	// using forward slashes this will remove them
	j, err := strconv.Unquote(j)
	if err != nil {
		return "", errors.Wrap(err, "could not unquote jsontag name")
	}

	// parsing the jsontag using fatih arslan's library
	tags, err := structtag.Parse(j)
	if err != nil {
		return "", errors.Wrap(err, "could not parse jsontag")
	}

	// as of now we assume that we have only one tag
	// in future if there is a use case where multiple
	// tags needed to be handled we make change here
	if len(tags.Tags()) > 1 {
		return "", fmt.Errorf("more than one tag found")
	}
	return tags.Tags()[0].Name, nil
}

// Parses comments on top of struct fields and accordingly returns the
// description of the field name if any provided, if the field is a reference
// defined using 'k8s:' or 'ref:' and also returns if the field is optional
// by looking for line that has '+optional' mentioned
func ParseStructFieldComments(cg *ast.CommentGroup) (desc string, ref string, optional bool) {
	// if no comments are given above the field then just return blank
	// strings, also assume that the field is optional
	if cg == nil {
		return "", "", true
	}

	// iterate on each line of comment
	// each comment is of the format "// blah blah"
	for _, c := range cg.List {
		comment := c.Text
		// comment also has leading // that we need to get rid of
		// and then if any space given before that we also need to remove it
		comment = strings.TrimSpace(strings.TrimPrefix(comment, "//"))

		if strings.HasPrefix(comment, "+optional") {
			// if the field is has optional mentioned mark the boolean as true
			optional = true
		} else if strings.HasPrefix(comment, "ref:") || strings.HasPrefix(comment, "k8s:") {
			// if this is reference either mentioned using 'ref' or 'k8s'
			// we remove the leading 'ref' or 'k8s' and return rest
			ref = strings.TrimSpace(strings.Split(comment, ":")[1])
		} else {
			// if none of above special cases then this is normal description
			desc = desc + comment + " "
		}
	}
	return strings.TrimSpace(desc), strings.TrimSpace(ref), optional
}

// Parses comments on top of structs and accordingly returns the
// description of struct and the kedgeSpec key if any provided
func ParseStructComments(cg *ast.CommentGroup) (kedgeSpecKey, desc string) {
	// if no comments are given above struct then just return blank
	// strings, note the empty return because the way function is defined
	if cg == nil {
		return
	}
	// iterate on each line of comment
	// each comment is of the format "// blah blah"
	for _, c := range cg.List {
		comment := c.Text
		// comment also has leading // that we need to get rid of
		// and then if any space given before that we also need to remove it
		comment = strings.TrimSpace(strings.TrimPrefix(comment, "//"))

		// check if the comment line specifies the key of 'kedgeSpec'
		// else it is normal comment so add it to the description
		if strings.HasPrefix(comment, "kedgeSpec:") {
			kedgeSpecKey = strings.TrimSpace(strings.Split(comment, ":")[1])
		} else {
			desc = desc + comment + " "
		}
	}
	return kedgeSpecKey, strings.TrimSpace(desc)
}

func PrintDefs(v interface{}) {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(string(b))
}
