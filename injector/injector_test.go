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

func NewStructC(a *StructA, iA InterfaceA, b StructB) *StructC {

	return &StructC{
		Text: "Hello! I'm StructC",
	}
}

func TestInjector(t *testing.T) {
	t.Run("Test LoadObject", func(t *testing.T) {
		injector.Init()

		injector.Provide[*StructA](NewStructA)
		injector.Provide[*StructB](NewStructB)

		injector.Provide[InterfaceA](NewStructA)

		injector.Provide[*StructC](NewStructC)
		injector.Provide[StructB](NewStructB2)

		c := injector.Get[*StructC]()
		if c.Text != "Hello! I'm StructC" {
			t.Errorf("Expected %s, got %s", "Hello! I'm StructC", c.Text)
		}
	})
}
