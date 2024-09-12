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
	"encoding/json"
	"fmt"

	"github.com/itchyny/gojq"
	"github.com/lf-edge/ekuiper/contract/v2/api"
	"github.com/lf-edge/ekuiper/v2/pkg/ast"
)

func registerJqFunc() {
	builtins["jq"] = builtinFunc{
		fType: ast.FuncTypeScalar,
		exec: func(ctx api.FunctionContext, args []interface{}) (interface{}, bool) {
			if len(args) != 2 {
				return fmt.Errorf("expected 2 arguments, got %d", len(args)), false
			}

			var input interface{}
			switch v := args[0].(type) {
			case string:
				if v == "" {
					return nil, true
				}
				if err := json.Unmarshal([]byte(v), &input); err != nil {
					return fmt.Errorf("invalid JSON string: %v", err), false
				}
			case []interface{}, map[string]interface{}:
				input = v
			case nil:
				return nil, true
			default:
				return fmt.Errorf("first argument must be a JSON string, array, or object, but got %T", v), false
			}

			queryStr, ok := args[1].(string)
			if !ok {
				return fmt.Errorf("second argument must be a string"), false
			}

			query, err := gojq.Parse(queryStr)
			if err != nil {
				return fmt.Errorf("error parsing JQ query: %v", err), false
			}

			iter := query.Run(input)
			var results []interface{}
			for {
				v, ok := iter.Next()
				if !ok {
					break
				}
				if err, ok := v.(error); ok {
					return fmt.Errorf("error processing query: %v", err), false
				}
				results = append(results, v)
			}

			if len(results) == 0 {
				return nil, true
			}

			if len(results) == 1 {
				return results[0], true
			}

			resultJSON, err := json.Marshal(results)
			if err != nil {
				return fmt.Errorf("error marshaling result to JSON: %v", err), false
			}

			return string(resultJSON), true

		},
		val: func(_ api.FunctionContext, args []ast.Expr) error {
			return ValidateLen(2, len(args))
		},
		check: returnNilIfHasAnyNil,
	}
}
