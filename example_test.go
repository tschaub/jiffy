package jiffy_test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tschaub/jiffy"
)

func Example_basic() {
	str := `["hello", "JSON", "Expression"]`
	expression := &jiffy.Expression{}

	err := json.Unmarshal([]byte(str), expression)
	if err != nil {
		log.Fatal("unexpected error:", err)
	}

	fmt.Printf("operator: %s\n", expression.Operator)
	fmt.Printf("arguments: %v\n", expression.Arguments)

	// Output:
	// operator: hello
	// arguments: [JSON Expression]
}

func Example_nested() {
	str := `["or", [">", 10], ["<", 20]]`
	expression := &jiffy.Expression{}

	err := json.Unmarshal([]byte(str), expression)
	if err != nil {
		log.Fatal("unexpected error:", err)
	}

	fmt.Printf("operator: %s\n", expression.Operator)
	fmt.Printf("number of arguments: %d\n", len(expression.Arguments))

	first, ok := expression.Arguments[0].(*jiffy.Expression)
	if !ok {
		log.Fatalf("unexpected argument: %v", expression.Arguments[0])
	}
	fmt.Printf("first: %s %v\n", first.Operator, first.Arguments)

	second, ok := expression.Arguments[1].(*jiffy.Expression)
	if !ok {
		log.Fatalf("unexpected argument: %v", expression.Arguments[0])
	}
	fmt.Printf("second: %s %v\n", second.Operator, second.Arguments)

	// Output:
	// operator: or
	// number of arguments: 2
	// first: > [10]
	// second: < [20]
}

func Example_validation() {
	// a custom validator for validating a "+" operator
	expression := &jiffy.Expression{
		Validator: func(operator string, arguments []interface{}) error {
			switch operator {
			case "+":
				if len(arguments) != 2 {
					return fmt.Errorf("the + operator takes two arguments, got %d", len(arguments))
				}
				for i, v := range arguments {
					if _, ok := v.(float64); !ok {
						return fmt.Errorf("expected number for argument %d, got %#v", i, v)
					}
				}
				return nil
			default:
				return fmt.Errorf(`unsupported operator "%s"`, operator)
			}
		},
	}

	valid := `["+", 10, 32]`
	if err := json.Unmarshal([]byte(valid), expression); err != nil {
		log.Fatal("unexpected error:", err)
	}
	fmt.Printf("%s passes validation\n", valid)

	notEnoughArgs := `["+", 42]`
	if err := json.Unmarshal([]byte(notEnoughArgs), expression); err != nil {
		fmt.Printf("%s fails validation: %s\n", notEnoughArgs, err)
	}

	wrongArgType := `["+", 10, "oops"]`
	if err := json.Unmarshal([]byte(wrongArgType), expression); err != nil {
		fmt.Printf("%s fails validation: %s\n", wrongArgType, err)
	}

	unsupportedOperation := `["oops", 42]`
	if err := json.Unmarshal([]byte(unsupportedOperation), expression); err != nil {
		fmt.Printf("%s fails validation: %s\n", unsupportedOperation, err)
	}

	// Output:
	// ["+", 10, 32] passes validation
	// ["+", 42] fails validation: the + operator takes two arguments, got 1
	// ["+", 10, "oops"] fails validation: expected number for argument 1, got "oops"
	// ["oops", 42] fails validation: unsupported operator "oops"
}
