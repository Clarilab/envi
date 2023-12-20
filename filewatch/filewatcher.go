package filewatch

import (
	"errors"
	"fmt"
)

type (
	// ConfigMap is a map of configuration variables.
	ConfigMap map[string]string

	// TriggerChannel is a channel
	TriggerChannel chan<- struct{}

	// ErrChan is a channel that is used to receive errors.
	ErrChan <-chan error

	callbackFunc func() error

	closeFunc func() error
)

type watcherType string

const (
	watcherTypeJSON watcherType = "json"
	watcherTypeYAML watcherType = "yaml"
)

// A FileWatcher can be used to observe a file on the filesystem.
// The Watcher will detect changes in the file, update the global ConfigMap in the application
// and send a struct{} to the given TriggerChannels which can be used for example to implement a
// connection-renewal for the desired technology instance.
type FileWatcher struct {
	Loader
	closeFunc
	watcherType
	errChan         <-chan error
	prefix          string
	path            string
	triggerChannels []TriggerChannel
}

// Option is a function that can be used to configure the FileWatcher.
type Option func(*FileWatcher)

// WithPrefix is a function that can be used to set the prefix for the file paths.
func WithPrefix(prefix string) Option {
	return func(f *FileWatcher) {
		f.prefix = prefix
	}
}

// WithLoader is a function that can be used to set the Loader for the FileWatcher.
func WithLoader(loader Loader) Option {
	return func(f *FileWatcher) {
		f.Loader = loader
	}
}

// WithTriggerChannels is a function that can be used to set the TriggerChannels for the FileWatcher.
func WithTriggerChannels(triggerChannels ...TriggerChannel) Option {
	return func(f *FileWatcher) {
		f.triggerChannels = triggerChannels
	}
}

// NewJSONFileWatcher creates a new File-Watcher that observes json files.
// The prefix is optional and can be left empty.
// This is useful in case you have multiple Watchers observing multiple files, which contain the same keys.
// The prefix will be added to the key in the global ConfigMap.
func NewJSONFileWatcher(path string, options ...Option) *FileWatcher {
	fw := &FileWatcher{
		watcherType: watcherTypeJSON,
		path:        path,
	}

	for i := range options {
		options[i](fw)
	}

	return fw
}

// NewYAMLFileWatcher creates a new File-Watcher that observes yaml files.
// The prefix is optional and can be left empty.
// This is useful in case you have multiple File-Watchers observing multiple files, which contain the same keys.
// When specified the prefix will be added to the key in the ConfigMap.
func NewYAMLFileWatcher(path string, options ...Option) *FileWatcher {
	fw := &FileWatcher{
		watcherType: watcherTypeYAML,
		path:        path,
	}

	for i := range options {
		options[i](fw)
	}

	return fw
}

// ErrLoaderNotSet is returned when no loader is specified.
var ErrLoaderNotSet = errors.New("no loader is specified")

// Start starts the file-watcher. The config parameter is the ConfigMap
// that will be updated when the file-watcher detects changes.
func (f *FileWatcher) Start(config ConfigMap) error {
	const errMessage = "failed to start watcher: %w"

	var err error

	if f.Loader == nil {
		return fmt.Errorf(errMessage, ErrLoaderNotSet)
	}

	switch {
	case f.watcherType == watcherTypeYAML && f.prefix != "":
		err, f.closeFunc, f.errChan = f.LoadAndWatchYAMLFilePrefixed(f.prefix, f.path, callback(config, f.ToMap, f.triggerChannels))

	case f.watcherType == watcherTypeJSON && f.prefix != "":
		err, f.closeFunc, f.errChan = f.LoadAndWatchJSONFilePrefixed(f.prefix, f.path, callback(config, f.ToMap, f.triggerChannels))

	case f.watcherType == watcherTypeYAML:
		err, f.closeFunc, f.errChan = f.LoadAndWatchYAMLFile(f.path, callback(config, f.ToMap, f.triggerChannels))

	case f.watcherType == watcherTypeJSON:
		err, f.closeFunc, f.errChan = f.LoadAndWatchJSONFile(f.path, callback(config, f.ToMap, f.triggerChannels))
	}
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

// Close closes the file-watcher.
func (f *FileWatcher) Close() error {
	return f.closeFunc()
}

// ErrChan returns the file-watcher's error channel.
func (f *FileWatcher) ErrChan() ErrChan {
	return f.errChan
}

// SetLoader sets the underlying loader for the file-watcher.
func (f *FileWatcher) SetLoader(loader Loader) {
	if loader != nil {
		f.Loader = loader
	}
}

func callback(config ConfigMap, toMap func() map[string]string, triggerChannels []TriggerChannel) callbackFunc {
	return func() error {
		config = toMap()

		if len(triggerChannels) > 0 {
			for i := range triggerChannels {
				triggerChannels[i] <- struct{}{}
			}
		}

		return nil
	}
}
