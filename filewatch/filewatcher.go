package filewatch

import (
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
type Option func(*FileWatcher) error

// WithPrefix is a function that can be used to set the prefix for the file paths.
func WithPrefix(prefix string) Option {
	return func(f *FileWatcher) error {
		if prefix == "" {
			return ErrEmptyPrefix
		}

		f.prefix = prefix

		return nil
	}
}

// WithTriggerChannels is a function that can be used to set the TriggerChannels for the FileWatcher.
func WithTriggerChannels(triggerChannels ...TriggerChannel) Option {
	return func(f *FileWatcher) error {
		if len(triggerChannels) == 0 {
			return ErrNoTriggers
		}

		f.triggerChannels = triggerChannels

		return nil
	}
}

// NewJSONFileWatcher creates a new File-Watcher that observes json files.
// The prefix is optional and can be left empty.
// Setting the Prefix is useful in case you have multiple Watchers observing multiple files,
// which contain the same keys. The prefix will be added to the key in the global ConfigMap.
func NewJSONFileWatcher(path string, loader Loader, options ...Option) (*FileWatcher, error) {
	return newWatcher(path, watcherTypeJSON, loader, options...)
}

// NewYAMLFileWatcher creates a new File-Watcher that observes yaml files.
// Setting the Prefix is useful in case you have multiple Watchers observing multiple files,
// which contain the same keys. The prefix will be added to the key in the global ConfigMap.
func NewYAMLFileWatcher(path string, loader Loader, options ...Option) (*FileWatcher, error) {
	return newWatcher(path, watcherTypeYAML, loader, options...)
}

func newWatcher(path string, typ watcherType, loader Loader, options ...Option) (*FileWatcher, error) {
	const errMessage = "failed to create a new YAML-File-Watcher: %w"

	if path == "" {
		return nil, fmt.Errorf(errMessage, ErrNoPath)
	}

	if loader == nil {
		return nil, fmt.Errorf(errMessage, ErrLoaderNotSet)
	}

	fw := &FileWatcher{
		watcherType: typ,
		Loader:      loader,
		path:        path,
	}

	for i := range options {
		if err := options[i](fw); err != nil {
			return nil, fmt.Errorf(errMessage, err)
		}
	}

	return fw, nil
}

// Start starts the file-watcher. The config parameter is the ConfigMap
// that will be updated when the file-watcher detects changes.
func (f *FileWatcher) Start(config ConfigMap) error {
	const errMessage = "failed to start watcher: %w"

	var err error

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
