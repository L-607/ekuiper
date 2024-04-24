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

package mqtt

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	mockContext "github.com/lf-edge/ekuiper/v2/pkg/mock/context"
)

func TestSinkConfigure(t *testing.T) {
	tests := []struct {
		name           string
		input          map[string]interface{}
		expectedErr    error
		expectedAdConf *AdConf
	}{
		{
			name: "TLS Error",
			input: map[string]interface{}{
				"topic":         "testTopic3",
				"qos":           0,
				"retained":      false,
				"compression":   "",
				"privateKeyRaw": "MTIz",
				"server":        "123",
			},
			expectedErr: fmt.Errorf("tls: failed to find any PEM data in certificate input"),
		},
		{
			name: "Missing topic",
			input: map[string]interface{}{
				"qos":         1,
				"retained":    false,
				"compression": "zlib",
			},
			expectedErr: fmt.Errorf("mqtt sink is missing property topic"),
		},
		{
			name: "Invalid QoS",
			input: map[string]interface{}{
				"topic":       "testTopic",
				"qos":         3,
				"retained":    false,
				"compression": "gzip",
			},
			expectedErr: fmt.Errorf("invalid qos value %v, the value could be only int 0 or 1 or 2", 3),
		},
		{
			name: "Valid configuration with QoS 0 and no compression",
			input: map[string]interface{}{
				"topic":       "testTopic3",
				"qos":         0,
				"retained":    false,
				"compression": "",
				"server":      "123",
			},
			expectedErr: nil,
			expectedAdConf: &AdConf{
				Tpc:         "testTopic3",
				Qos:         0,
				Retained:    false,
				Compression: "",
				ResendTopic: "testTopic3",
			},
		},
		{
			name: "Valid configuration with QoS 1 and no retained",
			input: map[string]interface{}{
				"topic":       "testTopic4",
				"qos":         1,
				"retained":    false,
				"compression": "zlib",
				"server":      "123",
			},
			expectedErr: nil,
			expectedAdConf: &AdConf{
				Tpc:         "testTopic4",
				Qos:         1,
				Retained:    false,
				Compression: "zlib",
				ResendTopic: "testTopic4",
			},
		},
	}

	ctx := mockContext.NewMockContext("testsinkconfigure", "sink1")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MQTTSink{}
			err := ms.Provision(ctx, tt.input)
			if !reflect.DeepEqual(err, tt.expectedErr) {
				t.Errorf("\n Expected error: \t%v\n \t\t\tgot: \t%v", tt.expectedErr, err)
				return
			}
			if tt.expectedErr == nil {
				if !reflect.DeepEqual(ms.adconf, tt.expectedAdConf) {
					t.Errorf("\n Expected adConf: \t%v\n \t\t\tgot: \t%v", tt.expectedAdConf, ms.adconf)
					return
				}
			}
		})
	}
}

func TestValidateMQTTSinkConf(t *testing.T) {
	testcases := []struct {
		topic       string
		expectError bool
	}{
		{
			topic:       "/123/+",
			expectError: true,
		},
		{
			topic:       "/123/#",
			expectError: true,
		},
		{
			topic: "/123/",
		},
	}
	for _, tc := range testcases {
		err := validateMQTTSinkTopic(tc.topic)
		if tc.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}