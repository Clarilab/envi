package filewatch

// Loader is an interface that is used to load files and converts them to a ConfigMap.
type Loader interface {
	ToMap() map[string]string
	LoadAndWatchJSONFile(path string, callback ...func() error) (error, func() error, <-chan error)
	LoadAndWatchYAMLFile(path string, callbacks ...func() error) (error, func() error, <-chan error)
	LoadAndWatchJSONFilePrefixed(prefix, path string, callback ...func() error) (error, func() error, <-chan error)
	LoadAndWatchYAMLFilePrefixed(prefix, path string, callbacks ...func() error) (error, func() error, <-chan error)
}
