package jiffy

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name string
	exp  *Expression
	str  string
	err  error
}

var cases = []TestCase{
	{
		name: "comparison expression",
		exp: &Expression{
			Operator: ">",
			Arguments: []interface{}{
				&Expression{
					Operator:  "get",
					Arguments: []interface{}{"count"},
				},
				42.0,
			},
		},
		str: `[">", ["get", "count"], 42]`,
	},

	{
		name: "more args",
		exp: &Expression{
			Operator:  "in",
			Arguments: []interface{}{"property", 42.0, 10.0, 100.0},
		},
		str: `["in", "property", 42, 10, 100]`,
	},

	{
		name: "bad operator type (number)",
		str:  `[42, "oops"]`,
		err:  errors.New("expected a string operator, got 42"),
	},

	{
		name: "bad operator type (boolean)",
		str:  `[true, "oops"]`,
		err:  errors.New("expected a string operator, got true"),
	},

	{
		name: "bad operator type (null)",
		str:  `[null, "oops"]`,
		err:  errors.New("expected a string operator, got <nil>"),
	},

	{
		name: "nested expression",
		exp: &Expression{
			Operator: "any",
			Arguments: []interface{}{
				&Expression{
					Operator:  ">",
					Arguments: []interface{}{"property", 42.0},
				},
				&Expression{
					Operator:  "<=",
					Arguments: []interface{}{"property", 100.0},
				},
			},
		},
		str: `["any", [">", "property", 42], ["<=", "property", 100]]`,
	},

	{
		name: "missing operator",
		exp: &Expression{
			Arguments: []interface{}{"oops"},
		},
		err: errors.New("zero length operator name"),
	},

	{
		name: "nested expression missing operator",
		exp: &Expression{
			Operator: "any",
			Arguments: []interface{}{
				&Expression{
					Operator:  ">",
					Arguments: []interface{}{"property", 42.0},
				},
				&Expression{
					Arguments: []interface{}{"oops"},
				},
			},
		},
		err: errors.New("failed to marshal argument 1: json: error calling MarshalJSON for type *jiffy.Expression: zero length operator name"),
	},

	{
		name: "bool arguments",
		exp: &Expression{
			Operator:  "bool",
			Arguments: []interface{}{true, false},
		},
		str: `["bool", true, false]`,
	},

	{
		name: "nil arguments",
		exp: &Expression{
			Operator: "void",
		},
		str: `["void"]`,
	},

	{
		name: "map arguments",
		exp: &Expression{
			Operator: "complex",
			Arguments: []interface{}{
				map[string]interface{}{"foo": 42.0},
			},
		},
		str: `["complex", {"foo": 42}]`,
	},

	{
		name: "pass validator",
		exp: &Expression{
			Operator:  "pass",
			Arguments: []interface{}{42.0},
			Validator: func(operator string, arguments []interface{}) error {
				if operator != "pass" {
					return fmt.Errorf("unexpected operator passed to validator '%s'", operator)
				}
				if len(arguments) != 1 {
					return fmt.Errorf("unexpected arguments passed to validator %v", arguments)
				}
				if arguments[0].(float64) != 42 {
					return fmt.Errorf("unexpected arguments passed to validator %v", arguments)
				}
				return nil
			},
		},
		str: `["pass", 42]`,
	},

	{
		name: "fail validator",
		exp: &Expression{
			Operator:  "fail",
			Arguments: []interface{}{42.0},
			Validator: func(operator string, arguments []interface{}) error {
				return errors.New("fail validator")
			},
		},
		err: errors.New("fail validator"),
	},
}

func assertMarshals(t *testing.T, tc TestCase) {
	jsonBytes, err := tc.exp.MarshalJSON()
	if err != nil {
		require.NotNil(t, tc.err, "expected no error, got %v", err)
		assert.EqualError(t, err, tc.err.Error())
		assert.Nil(t, jsonBytes)
		return
	}

	require.Nil(t, tc.err, "expected error %v", tc.err)
	assert.JSONEq(t, tc.str, string(jsonBytes))
}

func assertUnmarshals(t *testing.T, tc TestCase) {
	jsonBytes := []byte(tc.str)
	expression := &Expression{}
	err := json.Unmarshal(jsonBytes, expression)
	if err != nil {
		require.NotNil(t, tc.err, "expected no error, got %v", err)
		assert.EqualError(t, err, tc.err.Error())
		return
	}

	require.Nil(t, tc.err, "expected error %v", tc.err)
	assertEqual(t, tc.exp, expression)
}

func assertEqual(t *testing.T, expected *Expression, actual *Expression) {
	assert.Equal(t, expected.Operator, actual.Operator, "operator mismatch")
	require.Len(t, actual.Arguments, len(expected.Arguments))

	for i, arg := range expected.Arguments {
		got := actual.Arguments[i]
		assert.Equal(t, arg, got, "argument %d mismatch", i)
	}
}

func TestMarshalExpression(t *testing.T) {
	for i, tc := range cases {
		// skip cases without an expression
		if tc.exp == nil {
			continue
		}
		tc := tc
		t.Run(fmt.Sprintf("%s (%d)", tc.name, i), func(t *testing.T) { assertMarshals(t, tc) })
	}
}

func TestUnmarshalExpression(t *testing.T) {
	for i, tc := range cases {
		// skip cases without a string
		if tc.str == "" {
			continue
		}
		tc := tc
		t.Run(fmt.Sprintf("%s (%d)", tc.name, i), func(t *testing.T) { assertUnmarshals(t, tc) })
	}
}

func TestValidate(t *testing.T) {
	invalid := &Expression{
		Arguments: []interface{}{42},
	}

	invalidErr := invalid.Validate()
	assert.NotNil(t, invalidErr)

	valid := &Expression{
		Operator: "void",
	}

	validErr := valid.Validate()
	assert.Nil(t, validErr)
}

func TestCustomValidator(t *testing.T) {
	validator := func(operator string, arguments []interface{}) error {
		if operator == "void" {
			if len(arguments) > 0 {
				return errors.New("expected no arguments for void")
			}
			return nil
		}
		if len(arguments) == 0 {
			return errors.New("expected some arguments")
		}
		return nil
	}

	validVoid := &Expression{
		Operator:  "void",
		Validator: validator,
	}
	validVoidErr := validVoid.Validate()
	if validVoidErr != nil {
		t.Errorf("expected to validate: %s", validVoidErr)
	}

	invalidVoid := &Expression{
		Operator:  "void",
		Arguments: []interface{}{42},
		Validator: validator,
	}
	invalidVoidErr := invalidVoid.Validate()
	if invalidVoidErr == nil {
		t.Error("expected not to validate")
	}

	valid := &Expression{
		Operator:  "ident",
		Arguments: []interface{}{42},
		Validator: validator,
	}
	validErr := valid.Validate()
	if validErr != nil {
		t.Errorf("expected to validate: %s", validErr)
	}

	invalid := &Expression{
		Operator:  "ident",
		Validator: validator,
	}
	invalidErr := invalid.Validate()
	if invalidErr == nil {
		t.Error("expected not to validate")
	}
}

func TestCustomValidatorUnmarshal(t *testing.T) {
	validator := func(operator string, arguments []interface{}) error {
		if operator == "void" {
			if len(arguments) > 0 {
				return errors.New("expected no arguments for void")
			}
			return nil
		}
		if len(arguments) == 0 {
			return errors.New("expected some arguments")
		}
		return nil
	}

	cases := map[string]error{
		`["void"]`:                                nil,
		`["add", 42, 100]`:                        nil,
		`["void", "oops"]`:                        errors.New("expected no arguments for void"),
		`["add"]`:                                 errors.New("expected some arguments"),
		`["or", ["void"], ["add", 2, 2]]`:         nil,
		`["or", ["void"], ["oops"]]`:              errors.New("arg 1 error: expected some arguments"),
		`["or", ["or", ["void"], ["add", 2, 2]]]`: nil,
		`["or", ["or", ["void"], ["oops"]]]`:      errors.New("arg 0 error: arg 1 error: expected some arguments"),
	}

	for str, expectedErr := range cases {
		err := json.Unmarshal([]byte(str), &Expression{Validator: validator})
		if err != nil {
			if expectedErr == nil {
				t.Errorf("expected '%s' to validate: %s", str, err)
				continue
			}
			if err.Error() != expectedErr.Error() {
				t.Errorf("unexpected error for '%s': %s", str, err)
			}
			continue
		}
		if expectedErr != nil {
			t.Errorf("expected '%s' not to validate", str)
		}
	}
}
