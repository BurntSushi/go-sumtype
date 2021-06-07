package main

//go-sumtype:decl NotSealedT

// TestNotSealed
type NotSealedT interface{} // want "interface 'NotSealedT' is not sealed"
