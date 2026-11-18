# `go-luacats`

Generate [LuaCATS][LuaCATS] (Lua Comments and Type System) docstrings from [Go][Go] code.

Uses Go's built-in `go/parser` and `go/ast` packages to read [Go][Go] source code and extract type definitions and output formatted [LuaCATS][LuaCATS] definitions.

For a Go source file:

```go
package main

// Adds two numbers
func MyFunc(a, b int) int {
    return a + b
}

// Represents a simple structure
type MyStruct struct {
    X int
    Y string
    Z bool
}

// Does something with the struct
func (m MyStruct) SomeLogic() int {
    // Struct Method Implementation
}
```

generates LuaCATS docstrings:

```lua
--- Adds two numbers
---@param a integer
---@param b integer
---@return integer c
function MyFunc(a, b) end

--- Represents a simple structure
---@class MyStruct
---@field X integer
---@field Y string
---@field Z boolean
local MyStruct = {}

--- Does something with the struct
---@return integer
function MyStruct:SomeLogic() end
```

---

## Installation

From a checkout of this repository:

```sh
go install ./cmd/go-luacats
```

## Usage

### Command Line

```sh
go-luacats <source> [target]
```

`source` can be a Go file or a directory. For a file, `target` is a `.lua` file; for a directory, it is an output directory.

```sh
go-luacats path/to/file.go
go-luacats path/to/file.go output.lua
go-luacats path/to/directory/ generated/
```

Without a target, output is written alongside the source.

Directory processing is non-recursive and generates one Lua file per Go file.

> [!CAUTION]
> Existing output files are overwritten!

### Library


```go
package main

import (
    "log"

    "github.com/Shresht7/go-luacats"
)

func main() {
    err := luacats.Generate("path/to/file.go", "output.lua")
    if err != nil {
        log.Fatal(err)
    }
}
```

Pass an empty target to use the default output location

---

## License

[MIT License](./LICENSE)

[Go]: https://go.dev/
[LuaCATS]: https://github.com/LuaCATS/
[go-parser]: https://pkg.go.dev/go/parser
[go-ast]: https://pkg.go.dev/go/ast
