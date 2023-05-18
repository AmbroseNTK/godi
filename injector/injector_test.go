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

func NewStructB(a *StructA) StructB {
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
		err := injector.LoadObject(NewStructA())
		if err != nil {
			t.Errorf("LoadObject() error = %v", err)
		}
		err = injector.LoadObject(NewStructB(NewStructA()))
		if err != nil {
			t.Errorf("LoadObject() error = %v", err)
		}
		objs := injector.GetObjects()
		if len(objs) != 2 {
			t.Errorf("GetObjects() error = %v", objs)
		}
	})

	t.Run("Test Load", func(t *testing.T) {
		injector.Init()
		injector.Load(NewStructA)
		injector.Load(NewStructB)
		injector.Load(NewStructC)
		injector.LoadObject(NewStructA())
		result := injector.Inject[StructC](false)
		if result == nil {
			t.Errorf("Load() error = %v", result)
		}
	})
}
