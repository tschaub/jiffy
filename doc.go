/*
Package jiffy provides utilities for serializing and deserializing JSON Expressions.

A JSON Expression is a JSON array with three constraints: 1) the array must have length 1 or
greater, 2) the first element in the array must be a non-zero length string, and 3) all other
elements in the array that are arrays must be JSON Expressions.

The first element in a JSON Expression is the operator, and any remaining elements are
arguments for that operator.  For example, a JSON Expression representing a logical
expression that matches all numbers between 10 and 20 could look like this:

	[
	  "or",
	  ["gt", 10],
	  ["lt", 20]
	]

Here the "or" operator gets two arguments — both JSON Expressions.  The "gt" operator
(greater than) gets a single argument of 10.  The "lt" operator (less than) gets a single
argument of 20.

Grammar

A JSON Expression is a subset of JSON with a grammar that follows these rules:

	expression = begin-array operator *( value-separator argument ) end-array

	operator = quotation-mark 1*char quotation-mark

	argument = false / null / true / object / number / string / expression

See the JSON grammar (https://tools.ietf.org/html/rfc8259) for a definition of the
rules for begin-array, value-separator, end-array, quotation-mark, char, false, null,
true, object, number, and string.
*/
package jiffy
