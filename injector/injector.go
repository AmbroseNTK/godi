package injector

import (
	"errors"
	"reflect"
)

var dependencyMap map[string]Constructor = make(map[string]Constructor)
var objectMap = make(map[string]*Token)

var constructorMap = make(map[string]reflect.Type)

type Constructor struct {
	Path       string
	ObjectName string
	Type       reflect.Type
	FuncObj    interface{}
	Params     []*Token
}

type Token struct {
	Path  string
	Name  string
	Type  reflect.Type
	Value reflect.Value
}

func Init() {
	dependencyMap = make(map[string]Constructor)
	objectMap = make(map[string]*Token)
}

func LoadObject(obj interface{}) error {
	t := reflect.TypeOf(obj)
	var name = ""
	var path = ""
	if t.Kind() == reflect.Ptr {
		name = t.Elem().Name()
		path = t.Elem().PkgPath()
		t = t.Elem()
	} else {
		name = t.Name()
		path = t.PkgPath()
	}
	objectMap[path+"."+name] = &Token{Type: t, Name: name, Path: path, Value: reflect.ValueOf(obj)}
	return nil
}

func GetImplementationOfInterface(t reflect.Type) *Token {
	for _, token := range objectMap {
		obj := token.Value.Interface()
		if reflect.TypeOf(obj).Implements(t) {
			return token
		}
	}
	return nil
}

func Load(t interface{}) error {
	// get type of t
	tt := reflect.TypeOf(t)
	if tt.Kind() == reflect.Ptr {
		tt = tt.Elem()
	}
	// check if t is a constructor
	if tt.Kind().String() != "func" {
		return errors.New("t must be a constructor")
	}

	if tt.NumOut() == 2 {
		// check if the second return value is an error
		if tt.Out(1).Name() != "error" {
			return errors.New("the second return value must be an error")
		}
	}
	if tt.NumOut() < 1 || tt.NumOut() > 2 {
		return errors.New("the number of return values must be 1 or 2")
	}
	var name = ""
	var path = ""
	if tt.Out(0).Kind() == reflect.Ptr {
		name = tt.Out(0).Elem().Name()
		path = tt.Out(0).Elem().PkgPath()
	} else {
		name = tt.Out(0).Name()
		path = tt.Out(0).PkgPath()
	}

	constructorMap[name] = tt
	// get name of return value
	println(name)
	// get params of t
	var params []*Token

	for i := 0; i < tt.NumIn(); i++ {
		param := tt.In(i)
		if param.Kind() == reflect.Ptr {
			param = param.Elem()
		}
		params = append(params, &Token{
			Path: param.PkgPath(),
			Name: param.Name(), Type: param})
	}
	dependencyMap[path+"."+name] = Constructor{Path: path, ObjectName: name, Type: tt, Params: params, FuncObj: t}

	return nil
}

type myGeneric[T any] struct{}

func createObjectByConstructor(constructor Constructor) *Token {
	objectParam := make([]reflect.Value, 0)
	// get params
	for _, param := range constructor.Params {
		// get object by name
		// check if param is interface
		if param.Type.Kind() == reflect.Interface {
			// get implementation of interface
			implementation := GetImplementationOfInterface(param.Type)
			if implementation == nil {
				panic("implementation not found")
			} else {
				objectParam = append(objectParam, implementation.Value)
			}
			// get constructor of implementation

		}
		if param.Type.Kind() == reflect.Struct {
			// get object by name
			obj, ok := objectMap[param.Path+"."+param.Name]
			if !ok {
				// try to get constructor
				paramConstructor, ok := constructorMap[param.Name]
				if !ok {
					panic("object not found: " + param.Path + "." + param.Name)
				}
				obj = createObjectByConstructor(dependencyMap[paramConstructor.PkgPath()+"."+paramConstructor.Name()])
			}
			objectParam = append(objectParam, obj.Value)
		}
	}

	// call constructor by name
	result := reflect.ValueOf(constructor.FuncObj).Call(objectParam)
	// check if constructor returns error
	if len(result) == 2 {
		err := result[1]
		if !err.IsNil() {
			panic(err.Interface())
		}
	}
	return &Token{Type: constructor.Type, Name: constructor.ObjectName, Path: constructor.Path, Value: result[0]}
}

func Inject[T any](isGlobal bool) *T {
	// get name of type T
	generic := myGeneric[T]{}
	t := reflect.TypeOf(generic)
	typeName := t.Name()
	// substring
	typeName = typeName[10 : len(typeName)-1]
	println(typeName)
	// get constructor
	constructor, ok := dependencyMap[typeName]
	if !ok {
		panic("constructor not found")
	}
	println(constructor.Path + "." + constructor.ObjectName)
	// get in objectMap first
	if obj, ok := objectMap[constructor.Path+"."+constructor.ObjectName]; ok {
		return obj.Value.Interface().(*T)
	}
	objectParam := make([]reflect.Value, 0)
	// get params
	for _, param := range constructor.Params {
		// get object by name
		// check if param is interface
		if param.Type.Kind() == reflect.Interface {
			// get implementation of interface
			implementation := GetImplementationOfInterface(param.Type)
			if implementation == nil {
				panic("implementation not found")
			} else {
				objectParam = append(objectParam, implementation.Value)
			}
			// get constructor of implementation

		}
		if param.Type.Kind() == reflect.Struct {
			// get object by name
			// obj, ok := objectMap[param.Path+"."+param.Name]
			// if !ok {
			// 	// try to get constructor
			// 	paramConstructor, ok := constructorMap[param.Name]
			// 	if !ok {
			// 	panic("object not found: " + param.Path + "." + param.Name)
			// 	}

			// }
			obj := createObjectByConstructor(dependencyMap[param.Path+"."+param.Name])
			objectParam = append(objectParam, obj.Value)
		}
	}

	// call constructor by name
	result := reflect.ValueOf(constructor.FuncObj).Call(objectParam)
	// check if constructor returns error
	if len(result) == 2 {
		err := result[1]
		if !err.IsNil() {
			panic(err.Interface())
		}
	}
	if isGlobal {
		objectMap[constructor.Path+"."+constructor.ObjectName] = &Token{Type: constructor.Type, Name: constructor.ObjectName, Path: constructor.Path, Value: result[0]}
	}
	// parse result value to T
	return result[0].Interface().(*T)
}

func GetObjects() map[string]*Token {
	return objectMap
}
