# envi

## Installation

```shell
go get github.com/Clarilab/envi/v2
```

## Importing

```go
import "github.com/Clarilab/envi/v2"
```

## Available functions

```go
	// FromMap loads the given key-value pairs and loads them into the local map.
	FromMap(map[string]string)

	// LoadEnv loads the given keys from environment.
	LoadEnv(vars ...string)

	// LoadYAMLFilesFromEnvPaths loads yaml files from the paths in the given environment variables.
	LoadYAMLFilesFromEnvPaths(vars ...string) error 

	// LoadYAMLFilesFromEnvPaths loads json files from the paths in the given environment variables.
	LoadJSONFilesFromEnvPaths(vars ...string) error

	// LoadYAMLFilesFromEnvPaths loads the file content from the path in the given environment variable to the value of the given key.
	LoadFileFromEnvPath(key string, envPath string) error 

	// LoadFile loads a string value under given key from a file.
	LoadFile(key, filePath string) error

	// LoadJSON loads key-value pairs from one or many json blobs.
	LoadJSON(...[]byte) error

	// LoadJSONPrefixed loads key-value pairs from one or many json blobs
	// and prefixes the keys from the blobs with the given string.
	LoadJSONPrefixed(prefix string, blobs ...[]byte) error

	// LoadAndWatchJSONFile loads key-value pairs from a json file,
	// then watches that file and reloads it when it changes.
	// Accepts optional callback functions that are executed
	// after the file was reloaded. Returns and error when something
	// goes wrong. When no error is returned, returns a close function
	// that should be deferred in the calling function, and an error
	// channel where errors that occur during the file watching get sent.
	LoadAndWatchJSONFile(path string, callbacks ...func() error) (error, func() error, <-chan error)

	// LoadAndWatchJSONFilePrefixed works exactly as LoadAndWatchJSONFile,
	// except it prefixes the keys of the loaded variables with the given
	// string.
	LoadAndWatchJSONFilePrefixed(prefix, path string, callback ...func() error) (error, func() error, <-chan error)

	// LoadJSONFile loads key-value pairs from a json file.
	LoadJSONFile(path string) error

	// LoadJSONFilePrefixed loads key-value pairs from a json file
	// and prefixes the keys from the file with the given string.
	LoadJSONFilePrefixed(prefix, path string) error

	// LoadJSONFiles loads key-value pairs from one or more json files.
	LoadJSONFiles(...string) error

	// LoadYAML loads key-value pairs from one or many yaml blobs.
	LoadYAML(...[]byte) error

	// LoadYAMLPrefixed loads key-value pairs from one or many yaml blobs
	// and prefixes the keys from the blobs with the given string.
	LoadYAMLPrefixed(prefix string, blobs ...[]byte) error

	// LoadAndWatchYAMLFile loads key-value pairs from a yaml file,
	// then watches that file and reloads it when it changes.
	// Accepts optional callback functions that are executed
	// after the file was reloaded. Returns and error when something
	// goes wrong. When no error is returned, returns a close function
	// that should be deferred in the calling function, and an error
	// channel where errors that occur during the file watching get sent.
	LoadAndWatchYAMLFile(path string, callbacks ...func() error,) (error, func() error, <-chan error)

	// LoadAndWatchYAMLFilePrefixed works exactly as LoadAndWatchYAMLFile,
	// except it prefixes the keys of the loaded variables with the given
	// string.
	LoadAndWatchYAMLFilePrefixed(prefix, path string, callbacks ...func() error) (error, func() error, <-chan error)

	// LoadYAMLFile loads key-value pairs from a yaml file.
	LoadYAMLFile(path string) error 

	// LoadYAMLFilePrefixed loads key-value pairs from a yaml file
	// and prefixes the keys from the file with the given string.
	LoadYAMLFilePrefixed(prefix, path string) error

	// LoadYAMLFiles loads key-value pairs from one or more yaml files.
	LoadYAMLFiles(...string) error

	// EnsureVars checks, if all given keys have a non-empty value.
	EnsureVars(...string) error

	// ToEnv writes all key-value pairs to the environment.
	ToEnv()

	// ToMap returns a map, containing all key-value pairs.
	ToMap() map[string]string
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
