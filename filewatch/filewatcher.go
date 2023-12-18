package filewatch

import (
	"fmt"
)

type (
	// ConfigMap is a map of configuration variables.
	ConfigMap map[string]string

	// TriggerChannel is a channel that is used to
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

// NewJSONFileWatcher creates a new File-Watcher that observes json files.
// The prefix is optional and can be left empty.
// This is useful in case you have multiple Watchers observing multiple files, which contain the same keys.
// The prefix will be added to the key in the global ConfigMap.
func NewJSONFileWatcher(path, prefix string, triggerChannels ...TriggerChannel) *FileWatcher {
	return &FileWatcher{
		watcherType:     watcherTypeJSON,
		prefix:          prefix,
		path:            path,
		triggerChannels: triggerChannels,
	}
}

// NewYAMLFileWatcher creates a new File-Watcher that observes yaml files.
// The prefix is optional and can be left empty.
// This is useful in case you have multiple File-Watchers observing multiple files, which contain the same keys.
// When specified the prefix will be added to the key in the ConfigMap.
func NewYAMLFileWatcher(path, prefix string, triggerChannels ...TriggerChannel) *FileWatcher {
	return &FileWatcher{
		watcherType:     watcherTypeYAML,
		prefix:          prefix,
		path:            path,
		triggerChannels: triggerChannels,
	}
}

// Start starts the file-watcher. The config parameter is the ConfigMap
// that will be updated when the file-watcher detects changes.
func (f *FileWatcher) Start(config ConfigMap) error {
	return f.startWatcher(config)
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

func (f *FileWatcher) startWatcher(config ConfigMap) error {
	const errMessage = "failed to start watcher: %w"

	var err error
	var close closeFunc
	var errChan <-chan error

	switch {
	case f.watcherType == watcherTypeYAML && f.prefix != "":
		err, close, errChan = f.LoadAndWatchYAMLFilePrefixed(f.prefix, f.path, callback(config, f.ToMap, f.triggerChannels))

	case f.watcherType == watcherTypeJSON && f.prefix != "":
		err, close, errChan = f.LoadAndWatchJSONFilePrefixed(f.prefix, f.path, callback(config, f.ToMap, f.triggerChannels))

	case f.watcherType == watcherTypeYAML:
		err, close, errChan = f.LoadAndWatchYAMLFile(f.path, callback(config, f.ToMap, f.triggerChannels))

	case f.watcherType == watcherTypeJSON:
		err, close, errChan = f.LoadAndWatchJSONFile(f.path, callback(config, f.ToMap, f.triggerChannels))
	}
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	f.closeFunc = close
	f.errChan = errChan

	return nil
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
