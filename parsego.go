package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/structtag"
	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
)

type Injection struct {
	Target string
	Source string
}

// given a golang filename this function will parse the file and generate open api definition
func GenerateOpenAPIDefinitions(filename string) (spec.Definitions, []Injection, error) {
	// this has all the definitions which will be parsed from file
	defs := spec.Definitions(make(map[string]spec.Schema))
	// this stores all the mapping of what object fields to inject into what
	var mapping []Injection

	fset := token.NewFileSet() // positions are relative to fset
	// Parse the file also parse comments and add them to AST
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, mapping, errors.Wrapf(err, "could not read the go source code")
	}

	// iterate over all top-level declarations
	for _, decl := range node.Decls {
		// extract as generic declaration node
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		// iterate over all the specifications
		for _, s := range genDecl.Specs {
			// if there is a struct type it will be stored in strct
			strct, ok := TypeSpecToStruct(s)
			if !ok {
				continue
			}
			// function to parse struct
			m, err := ParseStruct(strct, genDecl, defs, fset)
			if err != nil {
				return nil, mapping, errors.Wrapf(err, "could not parse struct")
			}
			mapping = append(mapping, m...)
		}
	}

	log.Debugln("Mapping:")
	for _, s := range mapping {
		log.Debugln(s.Target, "-", s.Source)
	}
	PrintJSON(defs)

	return defs, mapping, nil
}

// Parses a struct object and creates a definition which is added with the key
// as specified in the comments of struct definition, also adds the keys as mentioned
// identifies the type of the fields and converts them into as needed by openapi
func ParseStruct(strct *ast.StructType, spc *ast.GenDecl, defs spec.Definitions, fset *token.FileSet) ([]Injection, error) {
	var mapping []Injection

	key, desc := ParseStructComments(spc.Doc)
	// Some fields are normal structs and are not part of
	// of schema, this can mostly happen when the struct is
	// embedded without redefining a key for it. e.g.
	// type PodSpecMod struct {
	if key == "" {
		return mapping, nil
	}
	CreateOpenAPIDefinition(key, desc, defs)

	// iterate all the fields of struct
	for _, sf := range strct.Fields.List {

		log.Debugln("Field name:", sf.Names)
		// To print using logrus we need to make the ast function
		// to write to bytes.Buffer and then extract string out of it
		var b bytes.Buffer
		ast.Fprint(&b, fset, sf, nil)
		log.Debug(b.String())

		// get the field name from the json tag
		name, err := JSONTagName(sf.Tag.Value)
		if err != nil {
			return mapping, errors.Wrapf(err, "name extraction from json tag error: %v", sf.Names)
		}

		// Find what is the type of struct field
		fieldtype, err := GetStructFieldType(sf.Type)
		if err != nil {
			return mapping, errors.Wrapf(err, "could not find the struct field type: %v", sf.Names)
		}

		// Parse comments written on top of struct field and then find the description
		// reference if any and see if the field is optional
		desc, ref, optional := ParseStructFieldComments(sf.Doc)

		// special cases of field types, after finding which we will do
		// some different processing rather than adding it to 'defs'
		switch fieldtype {
		case "":
			// this case will happen when we embed a struct in another
			// and if the struct is defined locally in same package
			// e.g.: PodSpecMod `json:",inline"`
			identifier, ok := sf.Type.(*ast.Ident)
			if !ok {
				continue
			}
			s, ok := TypeSpecToStruct(identifier.Obj.Decl)
			if !ok {
				continue
			}
			log.Debugln("Making a recursive call")
			m, err := ParseStruct(s, spc, defs, fset)
			if err != nil {
				return mapping, errors.Wrapf(err, "could not parse struct")
				continue
			}
			mapping = append(mapping, m...)
			continue
		case "selectorExpr":
			// This is case we have embedded a type from another package
			// so we just add it as mapping to so that we can inject the
			// definitions from that struct to our own definition
			s := Injection{Target: key, Source: ref}
			log.Debugf("add mapping {%q: %q}", s.Target, s.Source)
			mapping = append(mapping, s)
			continue
		}

		// for other types we just create schema and depending on the type
		// this will add necessary things
		schema, err := CreateSchema(fieldtype, desc, ref)
		if err != nil {
			return mapping, errors.Wrapf(err, "error creating schema: %v", sf.Names)
		}
		defs[key].Properties[name] = schema

		// also if the field is not optional then add it to the required list
		if !optional && name != "" {
			f := defs[key]
			f.Required = append(f.Required, name)
			defs[key] = f
		}
	}
	return mapping, nil
}

// Given two lists adds them, but only adds unique items
// Duplicates are removed, using map
func AddListUniqueItems(a []string, b []string) []string {
	var merger []string
	itemsMap := make(map[string]interface{})
	lists := [][]string{a, b}

	for _, list := range lists {
		for _, item := range list {
			itemsMap[item] = nil
		}
	}

	for k := range itemsMap {
		merger = append(merger, k)
	}
	return merger
}

// A TypeSpec node represents a type declaration (TypeSpec production).
// If given the object of that type returns StructType and boolean
// if the conversion went well
func TypeSpecToStruct(t interface{}) (*ast.StructType, bool) {
	ts, ok := t.(*ast.TypeSpec)
	if !ok {
		return nil, ok
	}
	strct, ok := ts.Type.(*ast.StructType)
	return strct, ok
}

// Schema the schema object allows the definition of input and output data types.
// This will be added to definitions we have
func CreateSchema(fieldtype, desc, ref string) (spec.Schema, error) {
	schema := spec.Schema{
		SchemaProps: spec.SchemaProps{
			Description: desc,
			Type:        spec.StringOrArray([]string{fieldtype}),
		},
	}
	switch fieldtype {
	case "object":
		schema.AdditionalProperties = &spec.SchemaOrBool{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray([]string{"string"}),
				},
			},
		}
	case "array":
		if ref != "" {
			refObj, err := CreateJSONRef(ref)
			if err != nil {
				return schema, errors.Wrapf(err, "error extracting name from json tag")
			}
			// an array of objects is list of objects being referred from somewhere
			// else so using
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
	case "starexpr":
		refObj, err := CreateJSONRef(ref)
		if err != nil {
			return schema, errors.Wrapf(err, "error extracting name from json tag")
		}
		schema.Ref = spec.Ref{refObj}
		// This was removed because all data is coming from that ref
		// there is no type called "starexpr" but this is more for knowing
		// that we need to refernce it directly
		schema.Type = nil
	}
	return schema, nil
}

// Creates a JSON reference type object from normal string
func CreateJSONRef(ref string) (jsonreference.Ref, error) {
	ref = "#/definitions/" + ref
	return jsonreference.New(ref)
}

// Given kedgeSpecKey and description for a struct this adds the entry
// for that particular key to defs, if not present already
func CreateOpenAPIDefinition(key, desc string, defs spec.Definitions) {
	if _, ok := defs[key]; !ok {
		defs[key] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: desc,
				Properties:  make(map[string]spec.Schema),
			},
		}
	}
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

func PrintJSON(v interface{}) {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		log.Fatalln(e)
	}
	log.Debugln(string(b))
}
