package jsonschema_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/mcombspangea/jsonschema/v5"
)

var ctx = context.Background()

func Example() {
	sch, err := jsonschema.Compile(ctx, "testdata/person_schema.json")
	if err != nil {
		log.Fatalf("%#v", err)
	}

	data, err := ioutil.ReadFile("testdata/person.json")
	if err != nil {
		log.Fatal(err)
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		log.Fatal(err)
	}

	if err = sch.Validate(ctx, v); err != nil {
		log.Fatalf("%#v", err)
	}
	// Output:
}

// Example_fromString shows how to load schema from string.
func Example_fromString() {
	schema := `{"type": "object"}`
	instance := `{"foo": "bar"}`

	sch, err := jsonschema.CompileString(ctx, "schema.json", schema)
	if err != nil {
		log.Fatalf("%#v", err)
	}

	var v interface{}
	if err := json.Unmarshal([]byte(instance), &v); err != nil {
		log.Fatal(err)
	}

	if err = sch.Validate(ctx, v); err != nil {
		log.Fatalf("%#v", err)
	}
	// Output:
}

// Example_fromStrings shows how to load schema from more than one string.
func Example_fromStrings() {
	c := jsonschema.NewCompiler()
	if err := c.AddResource("main.json", strings.NewReader(`{"$ref":"obj.json"}`)); err != nil {
		log.Fatal(err)
	}
	if err := c.AddResource("obj.json", strings.NewReader(`{"type":"object"}`)); err != nil {
		log.Fatal(err)
	}
	sch, err := c.Compile(ctx, "main.json")
	if err != nil {
		log.Fatalf("%#v", err)
	}

	var v interface{}
	if err := json.Unmarshal([]byte("{}"), &v); err != nil {
		log.Fatal(err)
	}

	if err = sch.Validate(ctx, v); err != nil {
		log.Fatalf("%#v", err)
	}
	// Output:
}

// Example_userDefinedFormat shows how to define 'odd-number' format.
func Example_userDefinedFormat() {
	jsonschema.Formats["odd-number"] = func(v interface{}) bool {
		switch v := v.(type) {
		case json.Number, float64, int, int32, int64:
			n, _ := strconv.ParseInt(fmt.Sprint(v), 10, 64)
			return n%2 != 0
		default:
			return true
		}
	}

	schema := `{
		"$schema": "http://json-schema.org/draft-07/schema",
		"type": "integer",
		"format": "odd-number"
	}`
	instance := 5

	sch, err := jsonschema.CompileString(ctx, "schema.json", schema)
	if err != nil {
		log.Fatalf("%#v", err)
	}

	if err = sch.Validate(ctx, instance); err != nil {
		log.Fatalf("%#v", err)
	}
	// Output:
}

// Example_userDefinedContent shows how to define
// "hex" contentEncoding and "application/xml" contentMediaType
func Example_userDefinedContent() {
	jsonschema.Decoders["hex"] = hex.DecodeString
	jsonschema.MediaTypes["application/xml"] = func(b []byte) error {
		return xml.Unmarshal(b, new(interface{}))
	}

	schema := `{
		"$schema": "http://json-schema.org/draft-07/schema",
		"type": "object",
		"properties": {
			"xml" : {
				"type": "string",
				"contentEncoding": "hex",
				"contentMediaType": "application/xml"
			}
		}
	}`
	instance := `{"xml": "3c726f6f742f3e"}`

	sch, err := jsonschema.CompileString(ctx, "schema.json", schema)
	if err != nil {
		log.Fatalf("%#v", err)
	}

	var v interface{}
	if err := json.Unmarshal([]byte(instance), &v); err != nil {
		log.Fatal(err)
	}

	if err = sch.Validate(ctx, v); err != nil {
		log.Fatalf("%#v", err)
	}
	// Output:
}

// Example_userDefinedLoader shows how to define custom schema loader.
//
// we are implementing a "map" protocol which servers schemas from
// go map variable.
func Example_userDefinedLoader() {
	var schemas = map[string]string{
		"main.json": `{"$ref":"obj.json"}`,
		"obj.json":  `{"type":"object"}`,
	}
	jsonschema.Loaders["map"] = func(url string) (io.ReadCloser, error) {
		schema, ok := schemas[strings.TrimPrefix(url, "map:///")]
		if !ok {
			return nil, fmt.Errorf("%q not found", url)
		}
		return ioutil.NopCloser(strings.NewReader(schema)), nil
	}

	sch, err := jsonschema.Compile(ctx, "map:///main.json")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	var v interface{}
	if err := json.Unmarshal([]byte("{}"), &v); err != nil {
		log.Fatal(err)
	}

	if err = sch.Validate(ctx, v); err != nil {
		log.Fatalf("%#v", err)
	}
	// Output:
}
