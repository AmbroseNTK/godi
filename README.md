# Godi: The most simplest Dependency Injection for Golang
## Diagram
![diagram](./diagrams/diagram.svg)
## Simple usage
```go
package main

import "github.com/AmbroseNTK/godi/injector"

func main() {
    injector.Init()
    // Provide your dependencies with their constructor
    injector.ProvideLazy[*StructA](NewStructA)
	injector.Provide[*InterfaceX](NewStructB)
    //...

    // Get your dependencies
    objA := injector.Get[*StructA]
}

```