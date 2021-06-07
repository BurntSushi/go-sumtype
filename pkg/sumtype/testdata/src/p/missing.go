package main

import "fmt"

//go-sumtype:decl T

type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type B struct {}
func (b *B) sealed() {}

type C struct {}
func (c *C) sealed() {}

func main() {
	// TestMissingNone
	switch T(nil).(type) {
	case *A, *B, *C:
	}

	// TestMissingOne
	switch T(nil).(type) { // want "exhaustiveness check failed for sum type 'T': missing cases for C"
	case *A:
	case *B:
	}

	// TestMissingTwo
	switch T(nil).(type) { // want "exhaustiveness check failed for sum type 'T': missing cases for A, C"
	case *B:
	}

	// TestMissingOneWithPanic
	switch T(nil).(type) { // want "exhaustiveness check failed for sum type 'T': missing cases for A"
	case *B:
	case *C:
	default:
		panic("unreachable")
	}

	// TestNoMissingDefault: default without panic thwarts exhaustiveness
	// checking
	switch T(nil).(type) {
	case *A:
	default:
		fmt.Println("legit catch all goes here")
	}
}