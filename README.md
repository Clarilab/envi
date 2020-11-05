# envi

## Installation
```shell
go get github.com/Clarilab/envi
```

## Importing
```go
import "github.com/Clarilab/envi"
```

## Features
```go
// LoadEnvVars loads the given Environment Variables.
// These are seperated into required and optional Vars.
func LoadEnvVars(required, optional []string) (loadedVars map[string]string, err error)
```
