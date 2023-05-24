package injector_test

import (
	"testing"

	"github.com/AmbroseNTK/godi/injector"
)

type StructA struct {
}

func NewStructA() *StructA {
	return &StructA{}
}

type StructB struct {
	A *StructA
}

func NewStructB(a *StructA) *StructB {
	return &StructB{A: a}
}
func NewStructB2(a *StructA) StructB {
	return StructB{A: a}
}

type InterfaceA interface {
	MethodA()
}

func (s *StructA) MethodA() {
	// Do something
}

type StructC struct {
	Text string
}

type StructZ struct {
	A *StructA
	X InterfaceA
}

func (z *StructZ) MethodZ() string {
	return "Hello! I'm StructZ"
}

func NewStructC(a *StructA, iA InterfaceA, b StructB) *StructC {

	return &StructC{
		Text: "Hello! I'm StructC",
	}
}

func TestInjector(t *testing.T) {
	t.Run("Test lazy provide", func(t *testing.T) {
		injector.Init()

		injector.ProvideLazy[*StructA](NewStructA)
		injector.ProvideLazy[*StructB](NewStructB)

		injector.ProvideLazy[InterfaceA](NewStructA)

		injector.ProvideLazy[*StructC](NewStructC)
		injector.ProvideLazy[StructB](NewStructB2)

		c := injector.Get[*StructC]()
		if c.Text != "Hello! I'm StructC" {
			t.Errorf("Expected %s, got %s", "Hello! I'm StructC", c.Text)
		}
	})
	t.Run("Test fields injection", func(t *testing.T) {
		injector.Init()
		injector.Provide[*StructA](NewStructA)
		injector.Provide[*StructB](NewStructB)
		z := injector.Inject[StructZ]()
		if z.MethodZ() != "Hello! I'm StructZ" {
			t.Errorf("Expected %s, got %s", "Hello! I'm StructZ", z.MethodZ())
		}

	})
}
