// Copyright 2025 EMQ Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package function

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/lf-edge/ekuiper/v2/internal/conf"
	"github.com/lf-edge/ekuiper/v2/internal/pkg/def"
	kctx "github.com/lf-edge/ekuiper/v2/internal/topo/context"
	"github.com/lf-edge/ekuiper/v2/internal/topo/state"
)

func TestFuncJq(t *testing.T) {
	t.Log("Starting TestFuncJq")
	contextLogger := conf.Log.WithField("rule", "testExec")
	ctx := kctx.WithValue(kctx.Background(), kctx.LoggerKey, contextLogger)
	tempStore, err := state.CreateStore("mockRule0", def.AtMostOnce)
	if err != nil {
		t.Fatalf("Failed to create temp store: %v", err)
	}
	t.Log("Temp store created")
	fctx := kctx.NewDefaultFuncContext(ctx.WithMeta("mockRule0", "test", tempStore), 2)
	t.Log("Function context created")
	registerJqFunc()
	t.Log("jq function registered")

	tests := []struct {
		name   string
		args   []interface{}
		result interface{}
		ok     bool
	}{
		{
			name: "valid JSON string with query",
			args: []interface{}{
				`{"name": "ekuiper", "version": "1.0"}`,
				".name",
			},
			result: "ekuiper",
			ok:     true,
		},
		{
			name: "valid JSON array with query",
			args: []interface{}{
				`[{"name": "ekuiper"}, {"name": "kuiper"}]`,
				".[0].name",
			},
			result: "ekuiper",
			ok:     true,
		},
		{
			name: "invalid JSON string",
			args: []interface{}{
				`{"name": "ekuiper", "version": "1.0"`,
				".name",
			},
			result: fmt.Errorf("invalid JSON string: %v", fmt.Errorf("unexpected end of JSON input")),
			ok:     false,
		},
		{
			name: "non-JSON first argument",
			args: []interface{}{
				123,
				".name",
			},
			result: fmt.Errorf("first argument must be a JSON string, array, or object, but got %T", 0),
			ok:     false,
		},
		{
			name: "non-string second argument",
			args: []interface{}{
				`{"name": "ekuiper"}`,
				123,
			},
			result: fmt.Errorf("second argument must be a string"),
			ok:     false,
		},
		{
			name: "valid JSON object with query returning multiple results",
			args: []interface{}{
				`{"name": "ekuiper", "version": "1.0"}`,
				".[]",
			},
			result: `["ekuiper","1.0"]`,
			ok:     true,
		},
		{
			name: "empty JSON object",
			args: []interface{}{
				`{}`,
				".name",
			},
			result: nil,
			ok:     true,
		},
		{
			name: "empty JSON array",
			args: []interface{}{
				`[]`,
				".[0]",
			},
			result: nil,
			ok:     true,
		},
		{
			name: "complex JSON object",
			args: []interface{}{
				`{"name": "ekuiper", "details": {"version": "1.0", "features": ["streaming", "analytics"]}}`,
				".details.features[1]",
			},
			result: "analytics",
			ok:     true,
		},
		{
			name: "query with filter",
			args: []interface{}{
				`[{"name": "ekuiper", "version": "1.0"}, {"name": "kuiper", "version": "2.0"}]`,
				".[] | select(.version == \"2.0\") | .name",
			},
			result: "kuiper",
			ok:     true,
		},
		{
			name: "multiple values from nested select",
			args: []interface{}{
				`[{"tags":["edge","iot"]},{"tags":["edge"]}]`,
				".[] | .tags[] | select(. == \"edge\")",
			},
			result: "[\"edge\",\"edge\"]",
			ok:     true,
		},
		{
			name: "object construction over multiple results",
			args: []interface{}{
				`{"items":[{"a":1},{"a":2}]}`,
				".items[] | {b: (.a * 2)}",
			},
			result: "[{\"b\":2},{\"b\":4}]",
			ok:     true,
		},
		{
			name: "index out of range returns nil",
			args: []interface{}{
				`[1,2]`,
				".[5]",
			},
			result: nil,
			ok:     true,
		},
		{
			name: "empty string input returns nil",
			args: []interface{}{
				"",
				".a",
			},
			result: nil,
			ok:     true,
		},
		{
			name: "map input with multiple emissions",
			args: []interface{}{
				map[string]interface{}{"x": map[string]interface{}{"y": 1}},
				".x | .y, .y",
			},
			result: "[1,1]",
			ok:     true,
		},
		{
			name: "numeric multiply",
			args: []interface{}{
				`{"num":3}`,
				".num*3",
			},
			result: float64(9),
			ok:     true,
		},
		{
			name: "string concat",
			args: []interface{}{
				`{"str":"343"}`,
				".str+\"3\"",
			},
			result: "3433",
			ok:     true,
		},
		{
			name: "indices on string",
			args: []interface{}{
				`"abcb"`,
				"indices(\"b\")",
			},
			result: []interface{}{1, 3},
			ok:     true,
		},
		{
			name: "map select filter to array",
			args: []interface{}{
				`[1,2,3,4]`,
				"map(select(.>2))",
			},
			result: []interface{}{float64(3), float64(4)},
			ok:     true,
		},
		{
			name: "del removes field",
			args: []interface{}{
				`{"name":{"firstname":"Void","lastname":"King"}}`,
				"del(.name.firstname)",
			},
			result: map[string]interface{}{"name": map[string]interface{}{"lastname": "King"}},
			ok:     true,
		},
		// Additional tests for documentation coverage
		{
			name: "array length",
			args: []interface{}{
				`[1,2,3,4]`,
				"length",
			},
			result: 4,
			ok:     true,
		},
		{
			name: "string length",
			args: []interface{}{
				`"hello"`,
				"length",
			},
			result: 5,
			ok:     true,
		},
		{
			name: "get keys",
			args: []interface{}{
				`{"name":"ekuiper","version":"1.0"}`,
				"keys",
			},
			result: []interface{}{"name", "version"},
			ok:     true,
		},
		{
			name: "has key true",
			args: []interface{}{
				`{"name":"ekuiper"}`,
				"has(\"name\")",
			},
			result: true,
			ok:     true,
		},
		{
			name: "has key false",
			args: []interface{}{
				`{"name":"ekuiper"}`,
				"has(\"age\")",
			},
			result: false,
			ok:     true,
		},
		{
			name: "array slicing",
			args: []interface{}{
				`[1,2,3,4,5]`,
				".[1:4]",
			},
			result: []interface{}{float64(2), float64(3), float64(4)},
			ok:     true,
		},
		{
			name: "flatten nested arrays",
			args: []interface{}{
				`[[1,2],[3,[4,5]]]`,
				"flatten",
			},
			result: []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)},
			ok:     true,
		},
		{
			name: "reverse array",
			args: []interface{}{
				`[1,2,3]`,
				"reverse",
			},
			result: []interface{}{float64(3), float64(2), float64(1)},
			ok:     true,
		},
		{
			name: "sort array",
			args: []interface{}{
				`[3,1,4,1,5]`,
				"sort",
			},
			result: []interface{}{float64(1), float64(1), float64(3), float64(4), float64(5)},
			ok:     true,
		},
		{
			name: "if-then-else adult",
			args: []interface{}{
				`{"age":25}`,
				"if .age >= 18 then \"adult\" else \"minor\" end",
			},
			result: "adult",
			ok:     true,
		},
		{
			name: "if-then-else minor",
			args: []interface{}{
				`{"age":15}`,
				"if .age >= 18 then \"adult\" else \"minor\" end",
			},
			result: "minor",
			ok:     true,
		},
		{
			name: "comparison operator",
			args: []interface{}{
				`{"a":5,"b":3}`,
				".a > .b",
			},
			result: true,
			ok:     true,
		},
		{
			name: "logical and",
			args: []interface{}{
				`{"a":5,"b":3}`,
				".a > 3 and .b < 5",
			},
			result: true,
			ok:     true,
		},
		{
			name: "add field",
			args: []interface{}{
				`{"name":"ekuiper"}`,
				".version = \"2.0\"",
			},
			result: map[string]interface{}{"name": "ekuiper", "version": "2.0"},
			ok:     true,
		},
		{
			name: "numeric addition",
			args: []interface{}{
				`{"a":10,"b":3}`,
				".a + .b",
			},
			result: float64(13),
			ok:     true,
		},
		{
			name: "numeric subtraction",
			args: []interface{}{
				`{"a":10,"b":3}`,
				".a - .b",
			},
			result: float64(7),
			ok:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running test case: %s", tt.name)
			f, ok := builtins["jq"]
			if !ok {
				t.Fatalf("jq function not registered")
			}
			result, ok := f.exec(fctx, tt.args)
			t.Logf("Test case result: %v, ok: %v", result, ok)
			if ok != tt.ok || !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Test case %s failed: got (%v, %v), expected (%v, %v)", tt.name, result, ok, tt.result, tt.ok)
			}
		})
	}
}
