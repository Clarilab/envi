# envi

## Installation

```shell
go get github.com/Clarilab/envi/v2
```

## Importing

```go
import "github.com/Clarilab/envi/v2"
```

## Features

```go
type Envi interface {
	// FromMap loads the given key-value pairs and loads them into the local map.
	FromMap(map[string]string)

	// LoadEnv loads the given keys from environment.
	LoadEnv(vars ...string)

	// LoadFile loads a string value under given key from a file.
	LoadFile(key, filePath string) error

	// LoadJSON loads key-value pairs from one or many json blobs.
	LoadJSON(...[]byte) error

	// LoadJSONFiles loads key-value pairs from one or more json files.
	LoadJSONFiles(...string) error

	// LoadYAML loads key-value pairs from one or many yaml blobs.
	LoadYAML(...[]byte) error

	// LoadYAMLFiles loads key-value pairs from one or more yaml files.
	LoadYAMLFiles(...string) error

	// EnsureVars checks, if all given keys have a non-empty value.
	EnsureVars(...string) error

	// ToEnv writes all key-value pairs to the environment.
	ToEnv()

	// ToMap returns a map, containing all key-value pairs.
	ToMap() map[string]string
}
```

### Examples

If you want to load key-values pairs from one or more json files into a map, you can use envi something like this.
```go
e := envi.NewEnvi()
err := e.LoadJSONFiles("./path/to/my/file.json", "./path/to/another/file.json")

myEnvVars := e.ToMap()

lookMom := myEnvVars["LOOK_MOM"]
```

To load environment variables something like this can be useful.
```go
e := envi.NewEnvi()
err := e.LoadEnv("HOME", "LOOK_MOM")

myEnvVars := e.ToMap()

lookMom := myEnvVars["LOOK_MOM"]
```

In many cases you want to ensure, that variables have a non empty value. This can be checked by using `EnsureVars()`.
```go
e := envi.NewEnvi()
err := e.LoadEnv("HOME", "LOOK_MOM")

err := e.EnsureVars("HOME")
if err != nil {
  // the error contains a list of missing vars.
  fmt.Println(err.MissingVars)

  // the Error() method prints the missing vars, too.
  fmt.Println(err
}
```

## Secretfile Example

A json file should look like this.
```json
{
    "SHELL": "csh",
    "PAGER": "more"
}
```

A basic yaml file like this.
```yaml
SHELL: bash
PAGER: less
```