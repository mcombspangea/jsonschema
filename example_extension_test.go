package jsonschema_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/mcombspangea/jsonschema"
)

var powerOfMeta = jsonschema.MustCompileString("powerOf.json", `{
	"properties" : {
		"powerOf": {
			"type": "integer",
			"exclusiveMinimum": 0
		}
	}
}`)

type powerOfCompiler struct{}

func (powerOfCompiler) Compile(cctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	if pow, ok := m["powerOf"]; ok {
		n, err := pow.(json.Number).Int64()
		return powerOfSchema(n), err
	}

	// nothing to compile, return nil
	return nil, nil
}

type powerOfSchema int64

func (s powerOfSchema) Validate(ctx context.Context, vctx jsonschema.ValidationContext, v interface{}) error {
	switch v.(type) {
	case json.Number, float64, int, int32, int64:
		pow := int64(s)
		n, _ := strconv.ParseInt(fmt.Sprint(v), 10, 64)
		for n%pow == 0 {
			n = n / pow
		}
		if n != 1 {
			e := vctx.Error(ctx, "powerOf", []interface{}{v, pow}, "%v not powerOf %v", v, pow)
			return e
		}
		return nil
	default:
		return nil
	}
}

func Example_extension() {
	ctx := context.Background()
	c := jsonschema.NewCompiler()
	c.RegisterExtension("powerOf", powerOfMeta, powerOfCompiler{})

	schema := `{"powerOf": 10}`
	instance := `100`

	if err := c.AddResource("schema.json", strings.NewReader(schema)); err != nil {
		log.Fatal(err)
	}
	sch, err := c.Compile(ctx, "schema.json")
	if err != nil {
		log.Fatalf("%#v", err)
	}

	if err = sch.Validate(ctx, strings.NewReader(instance)); err != nil {
		log.Fatalf("%#v", err)
	}
	// Output:
}
