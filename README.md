# envi

## Installation

```shell
go get github.com/Clarilab/envi/v3
```

## Importing

```go
import "github.com/Clarilab/envi/v3"
```

## Usage

### Config struct

To use envi, you have to create a struct which resembles your config structure that you want to parse.

```go
type Config struct {
	YamlFile    YAMLFile `env:"my-path-to.yaml" watch:"true"`
	JsonFile    JSONFile `env:"my-path-to.json" default:"./my-default-path.json" type:"json"`
	TextFile    TextFile `env:"my-path-to.txt" type:"text"`
	Environment string   `default:"prod"`
}

type TextFile struct {
	Value string `default:"foobar"`
}

type YAMLFile struct {
	StringField  string  `yaml:"STRING_FIELD"`
	IntField     int     `yaml:"INT_FIELD" default:"1337"`
	Int64Field   int64   `yaml:"INT_64_FIELD" required:"true" default:"1337"`
	BoolField    bool    `yaml:"BOOL_FIELD"`
	FloatField   float32 `yaml:"FLOAT_FIELD" required:"true"`
	Float64Field float64 `yaml:"FLOAT_64_FIELD" default:"3.1415926"`
}

// necessary func to enable file watching
func (y YAMLFile) OnChange() {
	// do something when the config filed changes
}

// necessary func to enable file watching
func (y YAMLFile) OnError(err error) {
	// do something when loading a config update fails
}

type JSONFile struct {
	Foo string `json:"FOO"`
	Bar int64  `json:"BAR"`
}
```

#### Available Tags

  - default: default value (supports file paths for files and standard data types bool, float32, float64, int32, int64, string)
  - env: environment variable name
  - type: describes the file type (json, yaml, text), defaults to yaml if omitted
  - required: indicates that the field is required, "envi.Load()" will return an error in this case
  - watch: indicates that the file should be watched for changes

#### File watcher

To watch for changes in config files, for example while using a vault, the underlying struct has to implement the envi.FileWatcher interface:

```go
type FileWatcher interface {
	OnChange()
	OnError(error)
}
```

### Load config

To load environment variables into your config:

```go
e := envi.New()
defer e.Close()

var myConfig Config
err := e.Load(&myConfig)
// some error handling as a good developer should do
```

Load loads all config files and environment variables into the input struct.
Supported types are JSON, YAML and text files, as well as strings on the struct root level.

If you want to watch a file for changes, the "watch" tag has to be set to true and the underlying struct
has to implement the envi.FileWatcher interface.

While using the "default" tag, the "env" tag can be omitted. If not omitted, the value from the
environment variable will be used.

When using the text file type, envi will try to load the file content into the first string field of that struct.
