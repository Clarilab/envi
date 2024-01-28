package filewatch

import (
	"fmt"
)

type (
	// TriggerChannel is a channel to send a signal when a file is changed.
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

// NewJSONFileWatcher creates a new FileWatcher that observes json files.
// Setting the Prefix is useful in case you have multiple Watchers observing multiple files,
// which contain the same keys. The prefix will be added to the key in the global ConfigMap.
func NewJSONFileWatcher(path string, loader Loader, options ...Option) (*FileWatcher, error) {
	return newWatcher(path, watcherTypeJSON, loader, options...)
}

// NewYAMLFileWatcher creates a new FileWatcher that observes yaml files.
// Setting the Prefix is useful in case you have multiple Watchers observing multiple files,
// which contain the same keys. The prefix will be added to the key in the global ConfigMap.
func NewYAMLFileWatcher(path string, loader Loader, options ...Option) (*FileWatcher, error) {
	return newWatcher(path, watcherTypeYAML, loader, options...)
}

func newWatcher(path string, typ watcherType, loader Loader, options ...Option) (*FileWatcher, error) {
	const errMessage = "failed to create a new YAML-FileWatcher: %w"

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

// Start starts the FileWatcher.
func (f *FileWatcher) Start() error {
	const errMessage = "failed to start watcher: %w"

	var err error

	switch {
	case f.watcherType == watcherTypeYAML && f.prefix != "":
		err, f.closeFunc, f.errChan = f.LoadAndWatchYAMLFilePrefixed(f.prefix, f.path, callback(f.triggerChannels))

	case f.watcherType == watcherTypeJSON && f.prefix != "":
		err, f.closeFunc, f.errChan = f.LoadAndWatchJSONFilePrefixed(f.prefix, f.path, callback(f.triggerChannels))

	case f.watcherType == watcherTypeYAML:
		err, f.closeFunc, f.errChan = f.LoadAndWatchYAMLFile(f.path, callback(f.triggerChannels))

	case f.watcherType == watcherTypeJSON:
		err, f.closeFunc, f.errChan = f.LoadAndWatchJSONFile(f.path, callback(f.triggerChannels))
	}
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

// Close closes the FileWatcher.
func (f *FileWatcher) Close() error {
	return f.closeFunc()
}

// ErrChan returns the FileWatcher's error channel.
func (f *FileWatcher) ErrChan() ErrChan {
	return f.errChan
}

func callback(triggerChannels []TriggerChannel) callbackFunc {
	return func() error {
		if len(triggerChannels) > 0 {
			for i := range triggerChannels {
				triggerChannels[i] <- struct{}{}
			}
		}

		return nil
	}
}
