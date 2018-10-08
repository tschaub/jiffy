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
	str := `["or", ["gt", 10], ["lt", 20]]`
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
	// first: gt [10]
	// second: lt [20]
}
