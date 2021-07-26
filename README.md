## Purpose

The start of a LISP interpreter having: tokenizer, parser, an AST tree and evaluator. Also an Environment.

Originally, for use in a talk at work.

## Limitations

- No unit tests
- Not very experienced in Golang, so the code is neither idiomatic
  nor very clean
- Handles only addition and multiplication for now

# Good Stuff

- Doesn't use a parser generator
- Shows how one can load a dictionary of primitives
- Tokenizer is also simple
- Conceptually based on https://norvig.com/lispy.html
