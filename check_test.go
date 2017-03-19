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
	tmpdir, prog := setupPackage(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(prog)
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
	tmpdir, prog := setupPackage(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(prog)
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
	tmpdir, prog := setupPackage(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(prog)
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
	tmpdir, prog := setupPackage(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(prog)
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
	tmpdir, prog := setupPackage(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(prog)
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
	tmpdir, prog := setupPackage(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(prog)
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
	tmpdir, prog := setupPackage(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(prog)
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
	tmpdir, prog := setupPackage(t, code)
	defer teardownPackage(t, tmpdir)

	errs := run(prog)
	if !assert.Len(t, errs, 1) {
		t.FailNow()
	}
	assert.Equal(t, "T", errs[0].(notInterfaceError).Decl.TypeName)
}

func missingNames(t *testing.T, err error) []string {
	if !assert.IsType(t, inexhaustiveError{}, err) {
		t.FailNow()
	}
	return err.(inexhaustiveError).Names()
}
