// Copyright 2023 EMQ Technologies Co., Ltd.
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

func TestJqFunc(t *testing.T) {
	t.Log("Starting TestJqFunc")
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
			result: fmt.Errorf("first argument must be a JSON string, array, or object"),
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
