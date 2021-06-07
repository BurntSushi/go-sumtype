package main

//go-sumtype:decl NotInterfaceT

// TestNotInterface
type NotInterfaceT struct{} // want "type 'NotInterfaceT' is not an interface"
