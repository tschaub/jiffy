package jiffy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Expression represents a JSON Expression.  An expression can be populated by
// calling the UnmarshalJSON method or with the json.Unmarshal function.  By default,
// unmarshaling validates that JSON conforms to the JSON Expression grammar.  An
// expression can be given a custom Validator function that will be called during
// unmarshalling.  The Validator function will be called with the expression's operator
// and arguments.
type Expression struct {
	Operator  string
	Arguments []interface{}
	Validator func(string, []interface{}) error // called with operator and arguments
}

// Validate determines if an expression is valid.  The only built-in requirement for validation
// is that the operator is a non-zero length string.  Any user-supplied Validator function will
// be run if there is a non-zero length operator string.
func (expression *Expression) Validate() error {
	if len(expression.Operator) == 0 {
		return errors.New("zero length operator name")
	}
	if expression.Validator != nil {
		return expression.Validator(expression.Operator, expression.Arguments)
	}
	return nil
}

// MarshalJSON returns the JSON encoding of an expression.  Expressions are validated before
// marshalling JSON.
func (expression *Expression) MarshalJSON() ([]byte, error) {
	validationErr := expression.Validate()
	if validationErr != nil {
		return nil, validationErr
	}

	var buffer bytes.Buffer
	buffer.WriteString(`[`)

	opBytes, opErr := json.Marshal(expression.Operator)
	if opErr != nil {
		return nil, fmt.Errorf("failed to marshal operator \"%s\": %s", expression.Operator, opErr)
	}

	_, writeErr := buffer.Write(opBytes)
	if writeErr != nil {
		return nil, fmt.Errorf("failed to write operator \"%s\": %s", expression.Operator, writeErr)
	}

	for i, argument := range expression.Arguments {
		buffer.WriteString(`,`)
		argBytes, argErr := json.Marshal(argument)
		if argErr != nil {
			return nil, fmt.Errorf("failed to marshal argument %d: %s", i, argErr)
		}
		_, writeErr := buffer.Write(argBytes)
		if writeErr != nil {
			return nil, fmt.Errorf("failed to write argument %d: %s", i, writeErr)
		}
	}

	buffer.WriteString(`]`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON creates an expression from JSON.  If the expression has a
// custom Validator function, this function will be called with the operator
// and arguments during unmarshalling.  Any nested expressions will acquire
// the same Validator function and must pass the same validation.
func (expression *Expression) UnmarshalJSON(data []byte) error {
	var parts []interface{}
	if partsErr := json.Unmarshal(data, &parts); partsErr != nil {
		return partsErr
	}

	return fromParts(parts, expression)
}

func getOperator(parts []interface{}) (string, error) {
	if len(parts) == 0 {
		return "", errors.New("expression must have an operator")
	}

	opInterface := parts[0]
	operator, ok := opInterface.(string)
	if !ok {
		return "", fmt.Errorf("expected a string operator, got %v", opInterface)
	}

	return operator, nil
}

func fromParts(parts []interface{}, expression *Expression) error {
	operator, opErr := getOperator(parts)
	if opErr != nil {
		return opErr
	}

	arguments := parts[1:]
	for i, arg := range arguments {
		nestedParts, ok := arg.([]interface{})
		if ok {
			nestedExpression := &Expression{
				Validator: expression.Validator,
			}
			nestedErr := fromParts(nestedParts, nestedExpression)
			if nestedErr != nil {
				return fmt.Errorf("arg %d error: %s", i, nestedErr)
			}
			arguments[i] = nestedExpression
		}
	}

	expression.Operator = operator
	expression.Arguments = arguments

	return expression.Validate()
}
