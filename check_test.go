package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMissingOne tests that we detect a single missing variant.
func TestMissingOne(t *testing.T) {
	code := `
package main

//go-sumtype:decl T

type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type B struct {}
func (b *B) sealed() {}

func main() {
	switch T(nil).(type) {
	case *A:
	}
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	if !assert.Len(t, errs, 1) {
		t.FailNow()
	}
	assert.Equal(t, []string{"B"}, missingNames(t, errs[0]))
}

// TestMissingTwo tests that we detect a two missing variants.
func TestMissingTwo(t *testing.T) {
	code := `
package main

//go-sumtype:decl T

type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type B struct {}
func (b *B) sealed() {}

type C struct {}
func (c *C) sealed() {}

func main() {
	switch T(nil).(type) {
	case *A:
	}
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	if !assert.Len(t, errs, 1) {
		t.FailNow()
	}
	assert.Equal(t, []string{"B", "C"}, missingNames(t, errs[0]))
}

// TestMissingOneWithPanic tests that we detect a single missing variant even
// if we have a trivial default case that panics.
func TestMissingOneWithPanic(t *testing.T) {
	code := `
package main

//go-sumtype:decl T

type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type B struct {}
func (b *B) sealed() {}

func main() {
	switch T(nil).(type) {
	case *A:
	default:
		panic("unreachable")
	}
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	if !assert.Len(t, errs, 1) {
		t.FailNow()
	}
	assert.Equal(t, []string{"B"}, missingNames(t, errs[0]))
}

// TestNoMissing tests that we correctly detect exhaustive case analysis.
func TestNoMissing(t *testing.T) {
	code := `
package main

//go-sumtype:decl T

type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type B struct {}
func (b *B) sealed() {}

type C struct {}
func (c *C) sealed() {}

func main() {
	switch T(nil).(type) {
	case *A, *B, *C:
	}
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	assert.Len(t, errs, 0)
}

// TestNoMissingDefault tests that even if we have a missing variant, a default
// case should thwart exhaustiveness checking.
func TestNoMissingDefault(t *testing.T) {
	code := `
package main

//go-sumtype:decl T

type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type B struct {}
func (b *B) sealed() {}

func main() {
	switch T(nil).(type) {
	case *A:
	default:
		println("legit catch all goes here")
	}
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	assert.Len(t, errs, 0)
}

// TestNotSealed tests that we report an error if one tries to declare a sum
// type with an unsealed interface.
func TestNotSealed(t *testing.T) {
	code := `
package main

//go-sumtype:decl T

type T interface {}

func main() {}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	if !assert.Len(t, errs, 1) {
		t.FailNow()
	}
	assert.Equal(t, "T", errs[0].(unsealedError).Decl.TypeName)
}

// TestNotFound tests that we report an error if one tries to declare a sum
// type that isn't defined.
func TestNotFound(t *testing.T) {
	code := `
package main

//go-sumtype:decl T

func main() {}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	if !assert.Len(t, errs, 1) {
		t.FailNow()
	}
	assert.Equal(t, "T", errs[0].(notFoundError).Decl.TypeName)
}

// TestNotInterface tests that we report an error if one tries to declare a sum
// type that doesn't correspond to an interface.
func TestNotInterface(t *testing.T) {
	code := `
package main

//go-sumtype:decl T

type T struct {}

func main() {}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	if !assert.Len(t, errs, 1) {
		t.FailNow()
	}
	assert.Equal(t, "T", errs[0].(notInterfaceError).Decl.TypeName)
}

// TestFormatterFriendly tests that we detect a decl even if the comment is
// prefixed with whitespace due to the formatter.
func TestFormatterFriendly(t *testing.T) {
	code := `
package main

// go-sumtype:decl T

func main() {}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	if !assert.Len(t, errs, 1) {
		t.FailNow()
	}
	assert.Equal(t, "T", errs[0].(notFoundError).Decl.TypeName)
}

// TestNoVariants tests that we report an error if one tries to declare a sum
// type with no variants.
func TestNoVariants(t *testing.T) {
	code := `
package main

//go-sumtype:decl T

type T interface { sealed() }

func main() {}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	if !assert.Len(t, errs, 1) {
		t.FailNow()
	}
	assert.Equal(t, "T", errs[0].(noVariantsError).Decl.TypeName)
}

// TestGenericInterfaceMissingOne tests that we detect a single missing variant
// for a generic interface.
func TestGenericInterfaceMissingOne(t *testing.T) {
	code := `
package main

//go-sumtype:decl Sealed

type Sealed[T any] interface { sealed() }

type Sealed1 struct{}
func (Sealed1) sealed() {}
var _ Sealed[struct{}] = Sealed1{}

type Sealed2 struct{}
func (Sealed2) sealed() {}
var _ Sealed[struct{}] = Sealed2{}

func Check[T any](s Sealed[T]) {
	switch s.(type) {
	case Sealed1:
	}
}

func main() {
	Check(Sealed2{})
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	if !assert.Len(t, errs, 1) {
		t.FailNow()
	}
	assert.Equal(t, []string{"Sealed2"}, missingNames(t, errs[0]))
}

// TestGenericInterfaceGenericVariantsMissingOne tests that we detect a single
// missing variant for a generic interface with generic variants.
func TestGenericInterfaceGenericVariantsMissingOne(t *testing.T) {
	code := `
package main

type Empty[T any] struct{}

//go-sumtype:decl Sealed

type Sealed[T any] interface { sealed(Empty[T]) }

type Sealed1[T any] struct{}
func (Sealed1[T]) sealed(Empty[T]) {}
var _ Sealed[struct{}] = Sealed1[struct{}]{}

type Sealed2[T any] struct{}
func (Sealed2[T]) sealed(Empty[T]) {}
var _ Sealed[struct{}] = Sealed2[struct{}]{}

func Check[T any](s Sealed[T]) {
	switch s.(type) {
	case Sealed1[T]:
	}
}

func main() {
	Check(Sealed2[struct{}]{})
}
`
	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	if !assert.Len(t, errs, 1) {
		t.FailNow()
	}
	assert.Equal(t, []string{"Sealed2"}, missingNames(t, errs[0]))
}

// TestNoMissingAlias tests that we correctly handle type aliases in
// exhaustiveness checking.
func TestNoMissingAlias(t *testing.T) {
	code := `
package main

//go-sumtype:decl T

type T interface { sealed() }

type A struct {}
func (a *A) sealed() {}

type A_Alias = A

func main() {
	switch T(nil).(type) {
	case *A_Alias:
	}
}
`

	tmpdir, pkgs := setupPackages(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(pkgs)
	assert.Len(t, errs, 0)
}

func missingNames(t *testing.T, err error) []string {
	if !assert.IsType(t, inexhaustiveError{}, err) {
		t.FailNow()
	}
	return err.(inexhaustiveError).Names()
}
