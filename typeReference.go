package main

type typeWrapper uint

const (
	twNone = iota
	twNonNull
	twList
)

type typeReference struct {
	referenceringType string
	referencingField  string
	referencedType    string
	typeWrapper       typeWrapper
}
