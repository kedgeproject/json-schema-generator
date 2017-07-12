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
		fieldtype := GetOpenAPIType(sf.Type)
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
		desc, ref, optional := GetStructFieldDesc(sf.Doc)

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

func GetOpenAPIType(g interface{}) string {
	t, ok := g.(*ast.Ident)
	if ok {
		switch t.Name {
		case "string":
			return "string"
		default:
			//panic(fmt.Sprintf("no type found: %s", t.Name))
			return ""
		}
	}

	mt, ok := g.(*ast.MapType)
	if ok {
		k, ok1 := mt.Key.(*ast.Ident)
		v, ok2 := mt.Value.(*ast.Ident)
		if ok1 && ok2 {
			if k.Name == "string" && v.Name == "string" {
				return "object"
			} else {
				panic("type not found")
			}
		} else {
			panic("type not found")
		}
	}

	_, ok = g.(*ast.ArrayType)
	if ok {
		return "array"
	}

	// api_v1.PersistentVolumeClaimSpec `json:",inline"`
	_, ok = g.(*ast.SelectorExpr)
	if ok {
		return "selectorExpr"
	}

	// ConfigMapRef *ConfigMapEnvSource `json:"configMapRef,omitempty"`
	_, ok = g.(*ast.StarExpr)
	if ok {
		return "starexpr"
	}

	panic(fmt.Sprintf("no type found"))
	return ""
}

func GetStructFieldDesc(cg *ast.CommentGroup) (desc string, ref string, optional bool) {
	if cg == nil {
		return "", "", true
	}

	for _, c := range cg.List {
		comment := c.Text
		comment = strings.TrimSpace(strings.TrimPrefix(comment, "//"))
		if strings.HasPrefix(comment, "+optional") {
			optional = true
		} else if strings.HasPrefix(comment, "ref:") || strings.HasPrefix(comment, "k8s:") {
			ref = strings.TrimSpace(strings.Split(comment, ":")[1])
		} else {
			desc = desc + comment + " "
		}
	}
	return strings.TrimSpace(desc), strings.TrimSpace(ref), optional
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
