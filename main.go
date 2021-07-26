package main

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

// ===========================
// Parsing stuff
// ===========================
type Tokens []string

func tokenize(chars string) Tokens {
	splitFn := func(c rune) bool {
		return c == ' '
	}
	replacer := strings.NewReplacer("(", " ( ", ")", " ) ")

	return strings.FieldsFunc(replacer.Replace(chars), splitFn)
}

func parseAtom(source string) SExp {
	if floatValue, err := strconv.ParseFloat(source, 64); err == nil {
		return AtomNumber{
			numericValue: floatValue,
		}
	}
	return AtomString{
		lexicalValue: source,
	}
}

func parseTokens(tokens Tokens, curr int) (SExp, int) {
	if len(tokens) == 0 {
		return Nil{}, 0
	} else if len(tokens) == 1 {
		return parseAtom(tokens[0]), 1
	} else if tokens[curr] == "(" {
		consumed := 1
		curr += 1
		head := &Node{}
		currHead := head
		for curr < len(tokens) {
			if tokens[curr] == "(" {
				retSexp, retConsumed := parseTokens(tokens, curr)
				currHead.val = retSexp
				currHead.next = &Node{}
				currHead = currHead.next
				curr += retConsumed
				consumed += retConsumed
			} else if tokens[curr] == ")" {
				curr += 1
				consumed += 1
				currHead.val = Nil{}
				currHead.next = nil
				return List{head: head}, consumed
			} else {
				currHead.val = parseAtom(tokens[curr])
				currHead.next = &Node{}
				currHead = currHead.next
				consumed += 1
				curr += 1
			}
		}
	}

	// If you got here you fell through
	panic("Malformed program, unclosed list")
}

// ===========================
// Evaluation and AST stuff
// ===========================

// SExp is the root of the AST
type SExp interface {
	eval(env *Env) SExp
}

// Env represents definitions for function symbols and atoms defined as
// values
type Env struct {
	localBindigs map[string]func(*Node, *Env) SExp
}

// While writing this I naively tried to return a pointer to the
// returned SEXp, but I hit the unaddressable operands issue
//
// https://utcc.utoronto.ca/~cks/space/blog/programming/GoAddressableValues
//
// Rules to read a few times.
func evalute(sexp SExp, env *Env) SExp {
	switch sexpt := sexp.(type) {
	case AtomNumber:
		return sexpt.eval(env)
	case AtomString:
		return sexpt.eval(env)
	case List:
		return sexpt.eval(env)
	case Nil:
		return sexpt.eval(env)
	default:
		log.Printf("unexpected type %T", sexpt)
		panic("Unknown type being evaluated")
	}
}

// Nil is the thing that is nothing
type Nil struct{}

func (n Nil) eval(env *Env) SExp {
	return n
}

// AtomString is the thing for initial values of Atoms in code but then
// gets evaluated later when needed. Sometimes a string is just a string
type AtomString struct {
	lexicalValue string
}

func (atom AtomString) eval(env *Env) SExp {
	// TODO: lookup value in environment return the eval of that
	return atom
}

// AtomNumber is the numberic twin of a numeric AtomString
type AtomNumber struct {
	numericValue float64
}

func (atom AtomNumber) eval(env *Env) SExp {
	// TODO: lookup value in environment return the eval of that
	return AtomNumber{numericValue: atom.numericValue}
}

// Node is for a classic List to represent lists. I know
// I could use slices, but this is kind of like car and cdr
// and is pretty easy to follow.
type Node struct {
	val  SExp
	next *Node
}

// List is a linear collection of Nodes
type List struct {
	head *Node
}

func (list List) eval(env *Env) SExp {
	if list.head == nil {
		return Nil{}
	}

	switch headSexpt := list.head.val.eval(env).(type) {
	case AtomString:
		if f, ok := env.localBindigs[headSexpt.lexicalValue]; ok {
			return f(list.head.next, env)
		} else {
			log.Printf("Not found in bindings %v", list.head)
			panic("Environmwnt is missing definition")
		}
	default:
		log.Printf("Lambda does not start with name of binding %v", list.head)
		panic("Unexpected type for lamda name")
	}
}

// QuotedSExp is used to treat the list as a lexical construct
type QuotedSExp struct {
	sexp SExp
}

func (quoted QuotedSExp) eval(env *Env) SExp {
	return quoted.sexp
}

func defaultEnv() *Env {
	// Native + operator - instead of some other low level implementation
	// with the idea being you build up on some basics like this and then
	// can define pure LISP values and code with set, defun etc.
	return &Env{
		localBindigs: map[string]func(*Node, *Env) SExp{
			"+": func(node *Node, env *Env) SExp {
				curr := node
				accum := 0.0

				for curr != nil {
					switch valt := curr.val.(type) {
					case AtomNumber:
						accum += valt.numericValue
					case AtomString:
						neval := valt.eval(env)
						switch nevalt := neval.(type) {
						case AtomNumber:
							accum += nevalt.numericValue
						default:
							log.Printf("Unsupported type %T for multiplcation operator", nevalt)
							panic("Multiplication of non-numeric values is not supported")
						}
					case List:
						neval := valt.eval(env)
						switch nevalt := neval.(type) {
						case AtomNumber:
							accum += nevalt.numericValue
						default:
							log.Printf("Unsupported type %T for addition operator", nevalt)
							panic("Addition of non-numeric values is not supported")
						}
					case Nil:
						if curr.next != nil {
							log.Printf("Malformed SExpression, Nil in middle of List")
							panic("Malformed SExpression, non-terminal Nil")
						}
					default:
						log.Printf("Unsupported value type %T for addition operator", valt)
						panic("Addition of unsupported value type")
					}
					curr = curr.next
				}

				return AtomNumber{
					numericValue: accum,
				}
			},
			"*": func(node *Node, env *Env) SExp {
				curr := node
				accum := 1.0

				for curr != nil {
					switch valt := curr.val.(type) {
					case AtomNumber:
						accum *= valt.numericValue
					case List:
						leval := valt.eval(env)
						switch levalt := leval.(type) {
						case AtomNumber:
							accum *= levalt.numericValue
						default:
							log.Printf("Unsupported type %T for multiplcation operator list processing", levalt)
							panic("Multiplication of non-numeric values is not supported")
						}
					case AtomString:
						neval := valt.eval(env)
						switch nevalt := neval.(type) {
						case AtomNumber:
							accum *= nevalt.numericValue
						default:
							log.Printf("Unsupported type %T for multiplcation operator", nevalt)
							panic("Multiplication of non-numeric values is not supported")
						}
					case Nil:
						if curr.next != nil {
							log.Printf("Malformed SExpression, Nil in middle of List")
							panic("Malformed SExpression, non-terminal Nil")
						}
					default:
						log.Printf("Unsupported value type %T for multiplication operator", valt)
						panic("Multiplication of unsupported value type")
					}
					curr = curr.next
				}

				return AtomNumber{
					numericValue: accum,
				}
			},
		},
	}
}

func main() {
	tokes1 := tokenize("(+ 10.5 -1.07)")
	fmt.Printf("tokes1 == %#v\n", tokes1)

	parsed1, consumed1 := parseTokens(tokes1, 0)
	fmt.Printf("%d parsed1  == %+v\n", consumed1, parsed1)
	result1 := evalute(parsed1, defaultEnv())

	fmt.Printf("result1 == %+v\n\ttype: %v\n", result1, reflect.TypeOf(result1))

	tokes2 := tokenize("(+ 10.5 -1.07 (* 10 (+ -1 1)))")
	fmt.Printf("tokes2 == %#v\n", tokes2)

	parsed2, consumed2 := parseTokens(tokes2, 0)
	fmt.Printf("%d parsed2  == %+v\n", consumed2, parsed2)
	result2 := evalute(parsed2, defaultEnv())

	fmt.Printf("result2 == %+v\n\ttype: %v\n", result2, reflect.TypeOf(result2))
}
