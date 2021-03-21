package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// TODO: parse it up
func tokenize(code string) []string {
	return strings.Split(strings.ReplaceAll(strings.ReplaceAll(code, "(", " ) "), ")", " ) "), " ")
}

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
	default:
		log.Printf("unexpected type %T", sexpt)
		panic("Unknown type being evaluated")
	}
}

// Nil is the thing that is nothing
type Nil struct{}

func (nil Nil) eval(env *Env) SExp {
	return nil
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

func main() {
	// (+ 10.5 -1.07)
	prog1 := List{
		head: &Node{
			val: AtomString{
				lexicalValue: "+",
			},
			next: &Node{
				val: AtomNumber{
					numericValue: 10.5,
				},
				next: &Node{
					val: AtomNumber{
						numericValue: -1.07,
					},
				},
			},
		},
	}
	// (* 13 10 10)
	prog2 := List{
		head: &Node{
			val: AtomString{
				lexicalValue: "*",
			},
			next: &Node{
				val: AtomNumber{
					numericValue: 13.0,
				},
				next: &Node{
					val: AtomNumber{
						numericValue: 10,
					},
					next: &Node{
						val: AtomNumber{
							numericValue: 10,
						},
					},
				},
			},
		},
	}

	// Native + operator - instead of some other low level implementation
	// with the idea being you build up on some basics like this and then
	// can define pure LISP values and code with set, defun etc.
	lispEnv := Env{
		localBindigs: map[string]func(*Node, *Env) SExp{
			"+": func(node *Node, env *Env) SExp {
				curr := node
				accum := 0.0

				for curr != nil {
					switch valt := curr.val.(type) {
					case AtomNumber:
						accum += valt.numericValue
					case List:
					case AtomString:
						neval := valt.eval(env)
						switch nevalt := neval.(type) {
						case AtomNumber:
							accum += nevalt.numericValue
						default:
							log.Printf("Unsupported type %T for addition operator", nevalt)
							panic("Addition of non-numeric values is not supported")
						}
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
					case AtomString:
						neval := valt.eval(env)
						switch nevalt := neval.(type) {
						case AtomNumber:
							accum *= nevalt.numericValue
						default:
							log.Printf("Unsupported type %T for addition operator", nevalt)
							panic("Addition of non-numeric values is not supported")
						}
					}
					curr = curr.next
				}

				return AtomNumber{
					numericValue: accum,
				}
			},
		},
	}

	result1 := evalute(prog1, &lispEnv)
	fmt.Printf("result1 == %+v\n\ttype: %v\n", result1, reflect.TypeOf(result1))

	result2 := evalute(prog2, &lispEnv)
	fmt.Printf("result2 == %+v\n\ttype: %v\n", result2, reflect.TypeOf(result2))
}
