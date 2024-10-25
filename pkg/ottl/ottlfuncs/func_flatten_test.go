// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ottlfuncs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
)

func Test_flatten(t *testing.T) {
	tests := []struct {
		name     string
		target   map[string]any
		prefix   ottl.Optional[string]
		depth    ottl.Optional[int64]
		expected map[string]any
		conflict bool
	}{
		{
			name: "simple",
			target: map[string]any{
				"name": "test",
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"name": "test",
			},
		},
		{
			name: "nested map",
			target: map[string]any{
				"address": map[string]any{
					"street": "first",
					"house":  int64(1234),
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"address.street": "first",
				"address.house":  int64(1234),
			},
		},
		{
			name: "conflicting map",
			target: map[string]any{
				"address": map[string]any{
					"street": map[string]any{
						"house": int64(1234),
					},
				},
				"address.street": map[string]any{
					"house": int64(1235),
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"address.street.house": int64(1235),
			},
		},
		{
			name: "conflicting slice",
			target: map[string]any{
				"address": map[string]any{
					"street": []any{"first"},
					"house":  int64(1234),
				},
				"address.street": []any{"second"},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"address.house":    int64(1234),
				"address.street.0": "second",
			},
		},
		{
			name: "conflicting map with nested slice",
			target: map[string]any{
				"address": map[string]any{
					"street": "first",
					"house":  int64(1234),
				},
				"address.street": "second",
				"occupants": []any{
					"user 1",
					"user 2",
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"address.street": "second",
				"address.house":  int64(1234),
				"occupants.0":    "user 1",
				"occupants.1":    "user 2",
			},
		},
		{
			name: "conflicting map with nested slice in conflicting item",
			target: map[string]any{
				"address": map[string]any{
					"street": map[string]any{
						"number": "first",
					},
					"house": int64(1234),
				},
				"address.street": map[string]any{
					"number": []any{"second", "third"},
				},
				"address.street.number": "fourth",
				"occupants": []any{
					"user 1",
					"user 2",
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"address.street.number":   "fourth",
				"address.house":           int64(1234),
				"address.street.number.0": "second",
				"address.street.number.1": "third",
				"occupants.0":             "user 1",
				"occupants.1":             "user 2",
			},
		},
		{
			name: "nested slice",
			target: map[string]any{
				"occupants": []any{
					"user 1",
					"user 2",
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"occupants.0": "user 1",
				"occupants.1": "user 2",
			},
		},
		{
			name: "combination",
			target: map[string]any{
				"name": "test",
				"address": map[string]any{
					"street": "first",
					"house":  int64(1234),
				},
				"occupants": []any{
					"user 1",
					"user 2",
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"name":           "test",
				"address.street": "first",
				"address.house":  int64(1234),
				"occupants.0":    "user 1",
				"occupants.1":    "user 2",
			},
		},
		{
			name: "deep nesting",
			target: map[string]any{
				"1": map[string]any{
					"2": map[string]any{
						"3": map[string]any{
							"4": "5",
						},
					},
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"1.2.3.4": "5",
			},
		},
		{
			name: "use prefix",
			target: map[string]any{
				"name": "test",
				"address": map[string]any{
					"street": "first",
					"house":  int64(1234),
				},
				"occupants": []any{
					"user 1",
					"user 2",
				},
			},
			prefix: ottl.NewTestingOptional[string]("app"),
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"app.name":           "test",
				"app.address.street": "first",
				"app.address.house":  int64(1234),
				"app.occupants.0":    "user 1",
				"app.occupants.1":    "user 2",
			},
		},
		{
			name: "max depth",
			target: map[string]any{
				"0": map[string]any{
					"1": map[string]any{
						"2": map[string]any{
							"3": "value",
						},
					},
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.NewTestingOptional[int64](2),
			expected: map[string]any{
				"0.1.2": map[string]any{
					"3": "value",
				},
			},
		},
		{
			name: "simple - conflict on",
			target: map[string]any{
				"name": "test",
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"name": "test",
			},
			conflict: true,
		},
		{
			name: "nested map - conflict on",
			target: map[string]any{
				"address": map[string]any{
					"street": "first",
					"house":  int64(1234),
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"address.street": "first",
				"address.house":  int64(1234),
			},
			conflict: true,
		},
		{
			name: "conflicting map - conflict on",
			target: map[string]any{
				"address": map[string]any{
					"street": map[string]any{
						"house": int64(1234),
					},
				},
				"address.street": map[string]any{
					"house": int64(1235),
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"address.street.house":   int64(1234),
				"address.street.house.0": int64(1235),
			},
			conflict: true,
		},
		{
			name: "conflicting slice - conflict on",
			target: map[string]any{
				"address": map[string]any{
					"street": []any{"first"},
					"house":  int64(1234),
				},
				"address.street": []any{"second"},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"address.street":   "first",
				"address.house":    int64(1234),
				"address.street.0": "second",
			},
			conflict: true,
		},
		{
			name: "conflicting map with nested slice - conflict on",
			target: map[string]any{
				"address": map[string]any{
					"street": "first",
					"house":  int64(1234),
				},
				"address.street": "second",
				"occupants": []any{
					"user 1",
					"user 2",
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"address.street":   "first",
				"address.house":    int64(1234),
				"address.street.0": "second",
				"occupants":        "user 1",
				"occupants.0":      "user 2",
			},
			conflict: true,
		},
		{
			name: "conflicting map with nested slice in conflicting item - conflict on",
			target: map[string]any{
				"address": map[string]any{
					"street": map[string]any{
						"number": "first",
					},
					"house": int64(1234),
				},
				"address.street": map[string]any{
					"number": []any{"second", "third"},
				},
				"address.street.number": "fourth",
				"occupants": []any{
					"user 1",
					"user 2",
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"address.street.number":   "first",
				"address.house":           int64(1234),
				"address.street.number.0": "second",
				"address.street.number.1": "third",
				"occupants":               "user 1",
				"occupants.0":             "user 2",
				"address.street.number.2": "fourth",
			},
			conflict: true,
		},
		{
			name: "nested slice - conflict on",
			target: map[string]any{
				"occupants": []any{
					"user 1",
					"user 2",
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"occupants":   "user 1",
				"occupants.0": "user 2",
			},
			conflict: true,
		},
		{
			name: "combination - conflict on",
			target: map[string]any{
				"name": "test",
				"address": map[string]any{
					"street": "first",
					"house":  int64(1234),
				},
				"occupants": []any{
					"user 1",
					"user 2",
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"name":           "test",
				"address.street": "first",
				"address.house":  int64(1234),
				"occupants":      "user 1",
				"occupants.0":    "user 2",
			},
			conflict: true,
		},
		{
			name: "deep nesting - conflict on",
			target: map[string]any{
				"1": map[string]any{
					"2": map[string]any{
						"3": map[string]any{
							"4": "5",
						},
					},
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"1.2.3.4": "5",
			},
			conflict: true,
		},
		{
			name: "use prefix - conflict on",
			target: map[string]any{
				"name": "test",
				"address": map[string]any{
					"street": "first",
					"house":  int64(1234),
				},
				"occupants": []any{
					"user 1",
					"user 2",
				},
			},
			prefix: ottl.NewTestingOptional[string]("app"),
			depth:  ottl.Optional[int64]{},
			expected: map[string]any{
				"app.name":           "test",
				"app.address.street": "first",
				"app.address.house":  int64(1234),
				"app.occupants":      "user 1",
				"app.occupants.0":    "user 2",
			},
			conflict: true,
		},
		{
			name: "max depth - conflict on",
			target: map[string]any{
				"0": map[string]any{
					"1": map[string]any{
						"2": map[string]any{
							"3": "value",
						},
					},
				},
			},
			prefix: ottl.Optional[string]{},
			depth:  ottl.NewTestingOptional[int64](2),
			expected: map[string]any{
				"0.1.2": map[string]any{
					"3": "value",
				},
			},
			conflict: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m := pcommon.NewMap()
			err := m.FromRaw(tt.target)
			assert.NoError(t, err)
			target := ottl.StandardPMapGetter[any]{
				Getter: func(_ context.Context, _ any) (any, error) {
					return m, nil
				},
			}

			exprFunc, err := flatten[any](target, tt.prefix, tt.depth, ottl.NewTestingOptional[bool](tt.conflict))
			assert.NoError(t, err)
			_, err = exprFunc(nil, nil)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, m.AsRaw())
		})
	}
}

func Test_flatten_bad_target(t *testing.T) {
	target := &ottl.StandardPMapGetter[any]{
		Getter: func(_ context.Context, _ any) (any, error) {
			return 1, nil
		},
	}
	exprFunc, err := flatten[any](target, ottl.Optional[string]{}, ottl.Optional[int64]{}, ottl.NewTestingOptional[bool](false))
	assert.NoError(t, err)
	_, err = exprFunc(nil, nil)
	assert.Error(t, err)
}

func Test_flatten_bad_depth(t *testing.T) {
	target := &ottl.StandardPMapGetter[any]{
		Getter: func(_ context.Context, _ any) (any, error) {
			return pcommon.NewMap(), nil
		},
	}
	_, err := flatten[any](target, ottl.Optional[string]{}, ottl.NewTestingOptional[int64](-1), ottl.NewTestingOptional[bool](false))
	assert.Error(t, err)
}
