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

func ProvideLazy[T any](constructor interface{}) error {
	t := reflect.TypeOf(constructor)
	if t.Kind() != reflect.Func {
		return errors.New("constructor must be a function")
	}
	if t.NumOut() == 0 {
		return errors.New("constructor must return at least one value")
	}
	ty := reflect.TypeOf((*T)(nil)).Elem()
	// check if ty is interface
	if ty.Kind() == reflect.Interface {
		// check if type T is an interface
		if t.Out(0).Implements(ty) {
			name := t.Out(0).String()
			constructors[name] = reflect.ValueOf(constructor)
			tName := reflect.TypeOf((*T)(nil)).Elem().String()
			constructors[tName] = reflect.ValueOf(constructor)
		} else {
			return errors.New("constructor must return a value that implements " + ty.String())
		}
	} else {
		// check if type T is the same as the return type of constructor
		if t.Out(0) == ty {
			name := t.Out(0).String()
			constructors[name] = reflect.ValueOf(constructor)
		} else {
			return errors.New("constructor must return a value of type " + ty.String())
		}
	}
	return nil
}

func Provide[T any](constructor interface{}) error {
	err := ProvideLazy[T](constructor)
	if err != nil {
		return err
	}
	Get[T]()
	return nil
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

func Inject[T any]() T {
	// create struct with fields injected
	t := reflect.TypeOf((*T)(nil)).Elem()
	name := t.String()
	if dep, ok := dependencies[name]; ok {
		return dep.(T)
	}
	// get fields of struct
	var fields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		fields = append(fields, t.Field(i))
	}
	// create struct with fields injected
	value := reflect.New(t).Elem()
	for _, field := range fields {
		dep, err := resolveDependency(field.Type)
		if err != nil {
			panic(err)
		}
		value.FieldByName(field.Name).Set(dep)
	}
	dependencies[name] = value.Interface()
	return value.Interface().(T)
}
