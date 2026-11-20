# `go-luacats`

Generate [LuaCATS][LuaCATS] (_Lua Comments and Type System_) docstrings from [Go][Go] code.

Uses Go's built-in `go/parser` and `go/ast` packages to read Go source code and extract type definitions and output formatted LuaCATS definitions.

> [!IMPORTANT]
> Currently, the generator emits annotations for exported `functions` and `methods`, exported `struct` types, and exported `struct fields`. 
> Other declarations are effectively ignored at this moment.

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

Requires Go 1.27.4 or newer.

Install the latest version:

```sh
go install github.com/Shresht7/go-luacats/cmd/go-luacats@latest
```

To run the CLI directly without installing first:

```sh
go run github.com/Shresht7/go-luacats/cmd/go-luacats@latest --help
```

To pin to a specific version, replace `@latest` with the desired version (e.g., `@v1.0.0`)

## Usage

### Command Line

```sh
go-luacats <source> [target] [--symbol package.Symbol ...]
```

Use `-h` or `--help` for help and `-v` or `--version` for version:

```sh
go-luacats --help
go-luacats --version
```

`source` can be a Go file, a directory, or `-` to read Go source from stdin.

Reading from stdin can be extremely useful when combined with the `go doc -src` command.

If no `target` is provided, generated LuaCATS definitions are written to `stdout`. For a directory, the output is combined into one document.

If `target` is provided:
- For a Go file, `target` is the output `.lua` file
- For a directory, `target` is the output directory; one `.lua` file per `.go` source file.

With `--symbol`, `source` is the Go module root and `target` is required. Each selector uses a package path relative to `source`
followed by a type name (e.g., `pkg.Type`). Repeat `--symbol` to select multiple types. The output directory mirrors the package paths, with
one `.lua` file per selected type. `--symbol` can appear before or after positional arguments.

```sh
go-luacats path/to/file.go
go-luacats path/to/file.go output.lua
go-luacats path/to/directory/
go-luacats path/to/directory/ generated/
go-luacats path/to/module generated/ --symbol internal/action.MyType --symbol internal/action.OtherType
go doc -src os.ReadFile | go-luacats -
```

When a target directory is provided, processing is **non-recursive** and generates one Lua file per top-level Go file.

> [!CAUTION]
> Existing output files are **overwritten**!

### Library


```go
package main

import (
    "log"

    "github.com/Shresht7/go-luacats"
)

func main() {
    err := luacats.GenerateFileFromFile("path/to/file.go", "output.lua")
    if err != nil {
        log.Fatal(err)
    }
}
```

For more control:
- Use `GenerateFromFile(sourceFile string, dst io.Writer)` when the source is a file and the destination can be any writer.
- Use `Generate(filename string, src io.Reader, dst io.Writer)` when you already have a source reader.

---

## License

[MIT License](./LICENSE)

[Go]: https://go.dev/
[LuaCATS]: https://github.com/LuaCATS/
[go-parser]: https://pkg.go.dev/go/parser
[go-ast]: https://pkg.go.dev/go/ast
