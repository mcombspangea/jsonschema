package jsonschema_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mcombspangea/jsonschema"
	_ "github.com/mcombspangea/jsonschema/httploader"
)

var testSuite = "testdata/JSON-Schema-Test-Suite@ab0b1ae"

var skipTests = map[string]map[string][]string{
	"TestDraft4/optional/unicode.json": {}, // golang regex works on ascii only
	"TestDraft4/optional/zeroTerminatedFloats.json": {
		"some languages do not distinguish between different types of numeric value": {}, // this behavior is changed in new drafts
	},
	"TestDraft4/optional/ecmascript-regex.json": {
		"ECMA 262 \\s matches whitespace": {
			"Line tabulation matches",                       // \s does not match vertical tab
			"latin-1 non-breaking-space matches",            // \s does not match unicode whitespace
			"zero-width whitespace matches",                 // \s does not match unicode whitespace
			"paragraph separator matches (line terminator)", // \s does not match unicode whitespace
			"EM SPACE matches (Space_Separator)",            // \s does not match unicode whitespace
		},
		"ECMA 262 \\S matches everything but whitespace": {
			"Line tabulation does not match",                       // \S matches unicode whitespace
			"latin-1 non-breaking-space does not match",            // \S matches unicode whitespace
			"zero-width whitespace does not match",                 // \S matches unicode whitespace
			"paragraph separator does not match (line terminator)", // \S matches unicode whitespace
			"EM SPACE does not match (Space_Separator)",            // \S matches unicode whitespace
		},
		"ECMA 262 regex escapes control codes with \\c and upper letter": {}, // \cX is not supported
		"ECMA 262 regex escapes control codes with \\c and lower letter": {}, // \cX is not supported
	},
	//
	"TestDraft6/optional/unicode.json": {}, // golang regex works on ascii only
	"TestDraft6/optional/ecmascript-regex.json": {
		"ECMA 262 \\s matches whitespace": {
			"Line tabulation matches",                       // \s does not match vertical tab
			"latin-1 non-breaking-space matches",            // \s does not match unicode whitespace
			"zero-width whitespace matches",                 // \s does not match unicode whitespace
			"paragraph separator matches (line terminator)", // \s does not match unicode whitespace
			"EM SPACE matches (Space_Separator)",            // \s does not match unicode whitespace
		},
		"ECMA 262 \\S matches everything but whitespace": {
			"Line tabulation does not match",                       // \S matches unicode whitespace
			"latin-1 non-breaking-space does not match",            // \S matches unicode whitespace
			"zero-width whitespace does not match",                 // \S matches unicode whitespace
			"paragraph separator does not match (line terminator)", // \S matches unicode whitespace
			"EM SPACE does not match (Space_Separator)",            // \S matches unicode whitespace
		},
		"ECMA 262 regex escapes control codes with \\c and upper letter": {}, // \cX is not supported
		"ECMA 262 regex escapes control codes with \\c and lower letter": {}, // \cX is not supported
	},
	//
	"TestDraft7/optional/unicode.json":             {}, // golang regex works on ascii only
	"TestDraft7/optional/format/idn-hostname.json": {}, // idn-hostname format is not implemented
	"TestDraft7/optional/format/idn-email.json":    {}, // idn-email format is not implemented
	"TestDraft7/optional/ecmascript-regex.json": {
		"ECMA 262 \\s matches whitespace": {
			"Line tabulation matches",                       // \s does not match vertical tab
			"latin-1 non-breaking-space matches",            // \s does not match unicode whitespace
			"zero-width whitespace matches",                 // \s does not match unicode whitespace
			"paragraph separator matches (line terminator)", // \s does not match unicode whitespace
			"EM SPACE matches (Space_Separator)",            // \s does not match unicode whitespace
		},
		"ECMA 262 \\S matches everything but whitespace": {
			"Line tabulation does not match",                       // \S matches unicode whitespace
			"latin-1 non-breaking-space does not match",            // \S matches unicode whitespace
			"zero-width whitespace does not match",                 // \S matches unicode whitespace
			"paragraph separator does not match (line terminator)", // \S matches unicode whitespace
			"EM SPACE does not match (Space_Separator)",            // \S matches unicode whitespace
		},
		"ECMA 262 regex escapes control codes with \\c and upper letter": {}, // \cX is not supported
		"ECMA 262 regex escapes control codes with \\c and lower letter": {}, // \cX is not supported
	},
	//
	"TestDraft2019/optional/unicode.json":             {}, // golang regex works on ascii only
	"TestDraft2019/optional/format/idn-hostname.json": {}, // idn-hostname format is not implemented
	"TestDraft2019/optional/format/idn-email.json":    {}, // idn-email format is not implemented
	"TestDraft2019/optional/ecmascript-regex.json": {
		"ECMA 262 \\s matches whitespace": {
			"Line tabulation matches",                       // \s does not match vertical tab
			"latin-1 non-breaking-space matches",            // \s does not match unicode whitespace
			"zero-width whitespace matches",                 // \s does not match unicode whitespace
			"paragraph separator matches (line terminator)", // \s does not match unicode whitespace
			"EM SPACE matches (Space_Separator)",            // \s does not match unicode whitespace
		},
		"ECMA 262 \\S matches everything but whitespace": {
			"Line tabulation does not match",                       // \S matches unicode whitespace
			"latin-1 non-breaking-space does not match",            // \S matches unicode whitespace
			"zero-width whitespace does not match",                 // \S matches unicode whitespace
			"paragraph separator does not match (line terminator)", // \S matches unicode whitespace
			"EM SPACE does not match (Space_Separator)",            // \S matches unicode whitespace
		},
		"ECMA 262 regex escapes control codes with \\c and upper letter": {}, // \cX is not supported
		"ECMA 262 regex escapes control codes with \\c and lower letter": {}, // \cX is not supported
	},
	//
	"TestDraft2020/optional/unicode.json":             {}, // golang regex works on ascii only
	"TestDraft2020/optional/format/idn-hostname.json": {}, // idn-hostname format is not implemented
	"TestDraft2020/optional/format/idn-email.json":    {}, // idn-email format is not implemented
	"TestDraft2020/optional/ecmascript-regex.json": {
		"ECMA 262 \\s matches whitespace": {
			"Line tabulation matches",                       // \s does not match vertical tab
			"latin-1 non-breaking-space matches",            // \s does not match unicode whitespace
			"zero-width whitespace matches",                 // \s does not match unicode whitespace
			"paragraph separator matches (line terminator)", // \s does not match unicode whitespace
			"EM SPACE matches (Space_Separator)",            // \s does not match unicode whitespace
		},
		"ECMA 262 \\S matches everything but whitespace": {
			"Line tabulation does not match",                       // \S matches unicode whitespace
			"latin-1 non-breaking-space does not match",            // \S matches unicode whitespace
			"zero-width whitespace does not match",                 // \S matches unicode whitespace
			"paragraph separator does not match (line terminator)", // \S matches unicode whitespace
			"EM SPACE does not match (Space_Separator)",            // \S matches unicode whitespace
		},
		"ECMA 262 regex escapes control codes with \\c and upper letter": {}, // \cX is not supported
		"ECMA 262 regex escapes control codes with \\c and lower letter": {}, // \cX is not supported
	},
}

func TestDraft4(t *testing.T) {
	testFolder(t, testSuite+"/tests/draft4", jsonschema.Draft4)
}

func TestDraft6(t *testing.T) {
	testFolder(t, testSuite+"/tests/draft6", jsonschema.Draft6)
}

func TestDraft7(t *testing.T) {
	testFolder(t, testSuite+"/tests/draft7", jsonschema.Draft7)
}

func TestDraft2019(t *testing.T) {
	testFolder(t, testSuite+"/tests/draft2019-09", jsonschema.Draft2019)
}

func TestDraft2020(t *testing.T) {
	testFolder(t, testSuite+"/tests/draft2020-12", jsonschema.Draft2020)
}

func TestExtra(t *testing.T) {
	t.Run("draft7", func(t *testing.T) {
		testFolder(t, "testdata/tests/draft7", jsonschema.Draft7)
	})
	t.Run("draft2020", func(t *testing.T) {
		testFolder(t, "testdata/tests/draft2020", jsonschema.Draft2020)
	})
}

type testGroup struct {
	Description string
	Schema      json.RawMessage
	Tests       []struct {
		Description string
		Data        json.RawMessage
		Valid       bool
		Skip        *string
	}
}

func TestMain(m *testing.M) {
	server := &http.Server{Addr: "localhost:1234", Handler: http.FileServer(http.Dir(testSuite + "/remotes"))}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()
	os.Exit(m.Run())
}

func testFolder(t *testing.T, folder string, draft *jsonschema.Draft) {
	fis, err := ioutil.ReadDir(folder)
	if err != nil {
		t.Fatal(err)
	}
	for _, fi := range fis {
		if fi.IsDir() {
			t.Run(fi.Name(), func(t *testing.T) {
				testFolder(t, path.Join(folder, fi.Name()), draft)
			})
			continue
		}
		if path.Ext(fi.Name()) != ".json" {
			continue
		}
		t.Run(fi.Name(), func(t *testing.T) {
			skip := skipTests[t.Name()]
			if skip != nil && len(skip) == 0 {
				t.Skip()
			}
			f, err := os.Open(path.Join(folder, fi.Name()))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			var tg []struct {
				Description string
				Schema      json.RawMessage
				Tests       []struct {
					Description string
					Data        interface{}
					Valid       bool
				}
			}
			dec := json.NewDecoder(f)
			dec.UseNumber()
			if err = dec.Decode(&tg); err != nil {
				t.Fatal(err)
			}
			for _, group := range tg {
				t.Run(group.Description, func(t *testing.T) {
					skip := skip[group.Description]
					if skip != nil && len(skip) == 0 {
						t.Skip()
					}
					c := jsonschema.NewCompiler()
					c.Draft = draft
					if strings.Index(folder, "optional") != -1 {
						c.AssertFormat = true
					}
					if err := c.AddResource("schema.json", bytes.NewReader(group.Schema)); err != nil {
						t.Fatal(err)
					}
					schema, err := c.Compile(ctx, "schema.json")
					if err != nil {
						t.Fatalf("%#v", err)
					}
					for _, test := range group.Tests {
						t.Run(test.Description, func(t *testing.T) {
							for _, desc := range skip {
								if test.Description == desc {
									t.Skip()
								}
							}
							err = schema.Validate(ctx, test.Data)
							valid := err == nil
							if !valid {
								if _, ok := err.(*jsonschema.ValidationError); ok {
									for _, line := range strings.Split(err.(*jsonschema.ValidationError).GoString(), "\n") {
										t.Logf("%s", line)
									}
								} else {
									t.Fatalf("got: %#v, want: *jsonschema.ValidationError", err)
								}
							}
							if test.Valid != valid {
								t.Fatalf("valid: got %v, want %v", valid, test.Valid)
							}
						})
					}
				})
			}
		})
	}
}

func TestMustCompile(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("panic expected")
			}
		}()
		jsonschema.MustCompile(ctx, "testdata/invalid_schema.json")
	})

	t.Run("valid", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Error("panic not expected")
				t.Log(r)
			}
		}()
		jsonschema.MustCompile(ctx, "testdata/person_schema.json")
	})
}

func TestInvalidSchema(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		if err := jsonschema.NewCompiler().AddResource("test.json", strings.NewReader("{")); err == nil {
			t.Error("error expected")
		} else {
			t.Logf("%v", err)
		}
	})

	t.Run("multiple json", func(t *testing.T) {
		if err := jsonschema.NewCompiler().AddResource("test.json", strings.NewReader("{}{}")); err == nil {
			t.Error("error expected")
		} else {
			t.Logf("%v", err)
		}
	})

	httpURL, httpsURL, cleanup := runHTTPServers()
	defer cleanup()
	invalidTests := []struct {
		description string
		url         string
	}{
		{"syntax error", "testdata/syntax_error.json"},
		{"missing filepath", "testdata/missing.json"},
		{"missing fileurl", toFileURL("testdata/missing.json")},
		{"missing httpurl", httpURL + "/missing.json"},
		{"missing httpsurl", httpsURL + "/missing.json"},
	}
	for _, test := range invalidTests {
		t.Run(test.description, func(t *testing.T) {
			if _, err := jsonschema.Compile(ctx, test.url); err == nil {
				t.Error("expected error")
			} else {
				t.Logf("%v", err)
			}
		})
	}

	type test struct {
		Description string
		Schema      json.RawMessage
		Fragment    string
	}
	data, err := ioutil.ReadFile("testdata/invalid_schemas.json")
	if err != nil {
		t.Fatal(err)
	}
	var tests []test
	if err = json.Unmarshal(data, &tests); err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			c := jsonschema.NewCompiler()
			url := "test.json"
			if err := c.AddResource(url, bytes.NewReader(test.Schema)); err != nil {
				t.Fatal(err)
			}
			if len(test.Fragment) > 0 {
				url += test.Fragment
			}
			if _, err = c.Compile(ctx, url); err == nil {
				t.Error("error expected")
			} else {
				t.Logf("%#v", err)
			}
		})
	}
}

func TestCompileURL(t *testing.T) {
	httpURL, httpsURL, cleanup := runHTTPServers()
	defer cleanup()

	validTests := []struct {
		schema, doc string
	}{
		//{"testdata/customer_schema.json#/0", "testdata/customer.json"},
		//{toFileURL("testdata/customer_schema.json") + "#/0", "testdata/customer.json"},
		//{httpURL + "/customer_schema.json#/0", "testdata/customer.json"},
		//{httpsURL + "/customer_schema.json#/0", "testdata/customer.json"},
		{toFileURL("testdata/empty schema.json"), "testdata/empty schema.json"},
		{httpURL + "/empty schema.json", "testdata/empty schema.json"},
		{httpsURL + "/empty schema.json", "testdata/empty schema.json"},
	}
	for i, test := range validTests {
		t.Run(test.schema, func(t *testing.T) {
			s, err := jsonschema.Compile(ctx, test.schema)
			if err != nil {
				t.Errorf("valid #%d: %v", i, err)
				return
			}
			f, err := os.Open(test.doc)
			if err != nil {
				t.Errorf("valid #%d: %v", i, err)
				return
			}
			err = s.Validate(ctx, f)
			_ = f.Close()
			if err != nil {
				t.Errorf("valid #%d: %v", i, err)
			}
		})
	}
}

func TestInvalidJsonTypeError(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	err := compiler.AddResource("test.json", strings.NewReader(`{ "type": "string"}`))
	if err != nil {
		t.Fatalf("addResource failed. reason: %v\n", err)
	}
	schema, err := compiler.Compile(ctx, "test.json")
	if err != nil {
		t.Fatalf("schema compilation failed. reason: %v\n", err)
	}
	v := struct{ name string }{"hello world"}
	err = schema.Validate(ctx, v)
	switch err.(type) {
	case jsonschema.InvalidJSONTypeError:
		// passed: struct is not valid json type
	default:
		t.Fatalf("got %v. want InvalidJSONTypeErr", err)
	}
}

func TestInfiniteLoopError(t *testing.T) {
	t.Run("compile", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		_, err := compiler.Compile(ctx, "testdata/loop-compile.json")
		if err == nil {
			t.Fatal("error expected")
		}
		switch err := err.(*jsonschema.SchemaError).Err.(type) {
		case jsonschema.InfiniteLoopError:
			suffix := "testdata/loop-compile.json#/$ref/$ref/not/$ref/allOf/0/$ref/anyOf/0/$ref/oneOf/0/$ref/dependencies/prop/$ref/dependentSchemas/prop/$ref/then/$ref/else/$ref"
			if !strings.HasSuffix(string(err), suffix) {
				t.Errorf("        got: %s", string(err))
				t.Errorf("want-suffix: %s", suffix)
			}
		default:
			t.Fatalf("got %#v. want InfiniteLoopTypeErr", err)
		}
	})
	t.Run("validate", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		schema, err := compiler.Compile(ctx, "testdata/loop-validate.json")
		if err != nil {
			t.Fatal(err)
		}
		err = schema.Validate(ctx, decodeString(t, `{"prop": 1}`))
		switch err := err.(type) {
		case jsonschema.InfiniteLoopError:
			suffix := "testdata/loop-validate.json#/$ref/$ref/not/$ref/allOf/0/$ref/anyOf/0/$ref/oneOf/0/$ref/dependencies/prop/$ref/dependentSchemas/prop/$ref/then/$ref/else/$dynamicRef/$ref"
			if !strings.HasSuffix(string(err), suffix) {
				t.Errorf("        got: %s", string(err))
				t.Errorf("want-suffix: %s", suffix)
			}
		default:
			t.Fatalf("got %#v. want InfiniteLoopTypeErr", err)
		}
	})
}

func TestExtractAnnotations(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()

		err := compiler.AddResource("test.json", strings.NewReader(`{
			"title": "this is title",
			"description": "this is description",
			"$comment": "this is comment",
			"format": "date-time",
			"examples": ["2019-04-09T21:54:56.052Z"],
			"readOnly": true,
			"writeOnly": true
		}`))
		if err != nil {
			t.Fatalf("addResource failed. reason: %v\n", err)
		}

		schema, err := compiler.Compile(ctx, "test.json")
		if err != nil {
			t.Fatalf("schema compilation failed. reason: %v\n", err)
		}

		if schema.Title != "" {
			t.Error("title should not be extracted")
		}
		if schema.Description != "" {
			t.Error("description should not be extracted")
		}
		if schema.Comment != "" {
			t.Error("comment should not be extracted")
		}
		if len(schema.Examples) != 0 {
			t.Error("examples should not be extracted")
		}
		if schema.ReadOnly {
			t.Error("readOnly should not be extracted")
		}
		if schema.WriteOnly {
			t.Error("writeOnly should not be extracted")
		}
	})

	t.Run("true", func(t *testing.T) {
		compiler := jsonschema.NewCompiler()
		compiler.ExtractAnnotations = true

		err := compiler.AddResource("test.json", strings.NewReader(`{
			"title": "this is title",
			"description": "this is description",
			"$comment": "this is comment",
			"format": "date-time",
			"examples": ["2019-04-09T21:54:56.052Z"],
			"readOnly": true,
			"writeOnly": true
		}`))
		if err != nil {
			t.Fatalf("addResource failed. reason: %v\n", err)
		}

		schema, err := compiler.Compile(ctx, "test.json")
		if err != nil {
			t.Fatalf("schema compilation failed. reason: %v\n", err)
		}

		if schema.Title != "this is title" {
			t.Errorf("title: got %q, want %q", schema.Title, "this is title")
		}
		if schema.Description != "this is description" {
			t.Errorf("description: got %q, want %q", schema.Description, "this is description")
		}
		if schema.Comment != "this is comment" {
			t.Errorf("$comment: got %q, want %q", schema.Comment, "this is comment")
		}
		if schema.Examples[0] != "2019-04-09T21:54:56.052Z" {
			t.Errorf("example: got %q, want %q", schema.Examples[0], "2019-04-09T21:54:56.052Z")
		}
		if !schema.ReadOnly {
			t.Error("readOnly should be extracted")
		}
		if !schema.WriteOnly {
			t.Error("writeOnly should be extracted")
		}
	})
}

func toFileURL(path string) string {
	path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	path = filepath.ToSlash(path)
	if runtime.GOOS == "windows" {
		path = "/" + path
	}
	u, err := url.Parse("file://" + path)
	if err != nil {
		panic(err)
	}
	return u.String()
}

// TestPanic tests https://github.com/mcombspangea/jsonschema/issues/18
func TestPanic(t *testing.T) {
	schema_d := `
	{
		"type": "object",
		"properties": {
		"myid": { "type": "integer" },
		"otype": { "$ref": "defs.json#someid" }
		}
	}
	`
	defs_d := `
	{
		"definitions": {
		"stt": {
			"$schema": "http://json-schema.org/draft-07/schema#",
			"$id": "#someid",
				"type": "object",
			"enum": [ { "name": "stainless" }, { "name": "zinc" } ]
		}
		}
	}
	`
	c := jsonschema.NewCompiler()
	c.Draft = jsonschema.Draft7
	if err := c.AddResource("schema.json", strings.NewReader(schema_d)); err != nil {
		t.Fatal(err)
	}
	if err := c.AddResource("defs.json", strings.NewReader(defs_d)); err != nil {
		t.Fatal(err)
	}

	if _, err := c.Compile(ctx, "schema.json"); err != nil {
		t.Fatal(err)
	}
}

func TestNonStringFormat(t *testing.T) {
	jsonschema.Formats["even-number"] = func(v interface{}) bool {
		switch v := v.(type) {
		case int:
			return v%2 == 0
		default:
			return false
		}
	}
	schema := `{"type": "integer", "format": "even-number"}`
	c := jsonschema.NewCompiler()
	c.AssertFormat = true
	if err := c.AddResource("schema.json", strings.NewReader(schema)); err != nil {
		t.Fatal(err)
	}
	s, err := c.Compile(ctx, "schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Validate(ctx, 5); err == nil {
		t.Fatal("error expected")
	}
	if err = s.Validate(ctx, 6); err != nil {
		t.Fatalf("%#v", err)
	}
}

func TestCompiler_LoadURL(t *testing.T) {
	const (
		base   = `{ "type": "string" }`
		schema = `{ "allOf": [{ "$ref": "base.json" }, { "maxLength": 3 }] }`
	)

	c := jsonschema.NewCompiler()
	c.LoadURL = func(s string) (io.ReadCloser, error) {
		switch s {
		case "map:///base.json":
			return ioutil.NopCloser(strings.NewReader(base)), nil
		case "map:///schema.json":
			return ioutil.NopCloser(strings.NewReader(schema)), nil
		default:
			return nil, errors.New("unsupported schema")
		}
	}
	s, err := c.Compile(ctx, "map:///schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Validate(ctx, "foo"); err != nil {
		t.Fatal(err)
	}
	if err = s.Validate(ctx, "long"); err == nil {
		t.Fatal("error expected")
	}
}

func TestFilePathSpaces(t *testing.T) {
	if _, err := jsonschema.Compile(ctx, "testdata/person schema.json"); err != nil {
		t.Fatal(err)
	}
}

func runHTTPServers() (httpURL, httpsURL string, cleanup func()) {
	tr := http.DefaultTransport.(*http.Transport)
	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{}
	}
	tr.TLSClientConfig.InsecureSkipVerify = true

	handler := http.FileServer(http.Dir("testdata"))
	httpServer := httptest.NewServer(handler)
	httpsServer := httptest.NewTLSServer(handler)

	return httpServer.URL, httpsServer.URL, func() {
		httpServer.Close()
		httpsServer.Close()
	}
}

func decodeString(t *testing.T, s string) interface{} {
	t.Helper()
	return decodeReader(t, strings.NewReader(s))
}

func decodeReader(t *testing.T, r io.Reader) interface{} {
	t.Helper()
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	var doc interface{}
	if err := decoder.Decode(&doc); err != nil {
		t.Fatal("invalid json:", err)
	}
	return doc
}

func TestUnevaluatedProperties(t *testing.T) {
	ctx := context.Background()
	rawSchema := `{
		"type": "object",
		"unevaluatedProperties": false
	}`
	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", strings.NewReader(rawSchema)); err != nil {
		t.Fatal(err)
	}
	schema, err := c.Compile(ctx, "schema.json")
	if err != nil {
		t.Fatalf("%#v", err)
	}
	err = schema.Validate(ctx, map[string]interface{}{
		"foo": "bar",
	})
	if err == nil {
		t.Fatal("Foo was an unevaluated property that went undetected")
	}
}
