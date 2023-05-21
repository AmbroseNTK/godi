package injector

import (
	"errors"
	"fmt"
	"reflect"
)

var dependencies = make(map[string]interface{})
var constructors = make(map[string]reflect.Value)

func Init() {
	dependencies = make(map[string]interface{})
	constructors = make(map[string]reflect.Value)
}

func Provide[T any](constructor interface{}) {
	t := reflect.TypeOf(constructor)
	if t.Kind() != reflect.Func {
		panic("Constructor must be a function")
	}
	if t.NumOut() == 0 {
		panic("Constructor must return at least one value")
	}
	name := t.Out(0).String()
	constructors[name] = reflect.ValueOf(constructor)
}

func resolveArguments(constructor reflect.Value) ([]reflect.Value, error) {
	var args []reflect.Value
	for j := 0; j < constructor.Type().NumIn(); j++ {
		argType := constructor.Type().In(j)
		// check if arg is interface
		if argType.Kind() == reflect.Interface {
			// find a dependency that implements the interface
			for _, dep := range dependencies {
				if reflect.TypeOf(dep).Implements(argType) {
					args = append(args, reflect.ValueOf(dep))
					break
				}
			}
			// if no dependency implements the interface, return error
			if len(args) < j+1 {
				return nil, fmt.Errorf("no dependency implements %s", argType.Name())
			}
		} else {
			dep, err := resolveDependency(argType)

			if err != nil {
				return nil, err
			}
			args = append(args, dep)
		}
	}
	return args, nil
}

func resolveDependency(t reflect.Type) (reflect.Value, error) {

	if t == nil {
		return reflect.Value{}, errors.New("invalid type")
	}

	name := ""

	// check if t is interface
	if t.Kind() == reflect.Interface {
		// find a dependency that implements the interface
		for _, dep := range dependencies {
			if reflect.TypeOf(dep).Implements(t) {
				return reflect.ValueOf(dep), nil
			}
		}
		// if no dependency implements the interface, return error
		return reflect.Value{}, fmt.Errorf("no dependency implements %s", t.Name())
	}

	// get full name of type
	name = t.String()

	if dep, ok := dependencies[name]; ok {
		return reflect.ValueOf(dep), nil
	}

	creating := make(map[string]bool)

	if _, ok := creating[name]; ok {
		return reflect.Value{}, fmt.Errorf("circular dependency detected for %s", name)
	}

	creating[name] = true
	constructor, ok := constructors[name]
	if !ok {
		return reflect.Value{}, fmt.Errorf("no constructor registered for %s", name)
	}

	args, err := resolveArguments(constructor)
	if err != nil {
		return reflect.Value{}, err
	}

	value := constructor.Call(args)[0]
	dependencies[name] = value.Interface()
	delete(creating, name)

	return value, nil
}

func Get[T any]() T {
	t := reflect.TypeOf((*T)(nil)).Elem()
	name := t.String()
	if dep, ok := dependencies[name]; ok {
		return dep.(T)
	}

	constructor, ok := constructors[name]
	if !ok {
		panic(fmt.Sprintf("No constructor registered for %s", name))
	}

	args, err := resolveArguments(constructor)
	if err != nil {
		panic(err)
	}

	value := constructor.Call(args)[0].Interface().(T)
	dependencies[name] = value

	return value
}
