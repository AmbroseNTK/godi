package injector

import (
	"errors"
	"fmt"
	"reflect"
)

var dependencies = make(map[string]interface{})
var constructors = make(map[string]reflect.Value)

// Init initializes the dependencies and constructors maps.
// Init must be called before any other function in this package.
func Init() {
	dependencies = make(map[string]interface{})
	constructors = make(map[string]reflect.Value)
}

// ProvideLazy registers a constructor function that returns a value of type T.
// The constructor function is not called until Get[T]() is called.
// If T is an interface, the constructor function must return a value that implements T.
// If T is not an interface, the constructor function must return a value of type T.
// The constructor function must return at least one value.
// The constructor function must be a function.
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

// Provide registers a constructor function that returns a value of type T.
// The constructor function is called immediately and the result is stored as a dependency.
// If T is an interface, the constructor function must return a value that implements T.
// If T is not an interface, the constructor function must return a value of type T.
// The constructor function must return at least one value.
// The constructor function must be a function.
// Provide is a shorthand for ProvideLazy followed by Get[T].
func Provide[T any](constructor interface{}) error {
	err := ProvideLazy[T](constructor)
	if err != nil {
		return err
	}
	Get[T]()
	return nil
}

// resolveArguments resolves the dependencies of a constructor function and returns them as a slice of reflect.Value.
// The constructor function is passed as a reflect.Value.
// If an argument of the constructor function is an interface, resolveArguments finds a dependency that implements the interface.
// If an argument of the constructor function is not an interface, resolveArguments finds the dependency by name.
// If a dependency cannot be found, resolveArguments returns an error.
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

// resolveDependency returns the dependency of a given type.
// The dependency is retrieved from the dependencies map.
// If the dependency is not found, an error is returned.
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

// Get returns an instance of the type specified by the type parameter T.
// If an instance has already been created, it is returned from the cache.
// Otherwise, a new instance is created using the registered constructor function.
// The constructor function is called with its dependencies resolved using reflection.
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

// Inject creates a new instance of the specified struct type with its fields injected with their respective dependencies.
// It returns the created instance.
func Inject[T any]() T {
	isPointer := false
	// create struct with fields injected
	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() == reflect.Ptr {
		isPointer = true
		t = t.Elem()
	}
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
	if isPointer {
		return value.Addr().Interface().(T)
	}
	return value.Interface().(T)
}
