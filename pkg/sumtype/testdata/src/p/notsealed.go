package main

//go-sumtype:decl NotSealedT

// TestNotSealed
type NotSealedT interface {} // want "not sealed"
