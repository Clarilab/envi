package filewatch_test

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Clarilab/envi/v2"
	"github.com/Clarilab/envi/v2/filewatch"
)

const (
	prefix   = "PRE_"
	userName = "USERNAME"
	password = "PASSWORD"

	keyUserName = prefix + userName
	keyPassword = prefix + password

	yamlFilePath = "./testdata/test-secrets.yaml"
	jsonFilePath = "./testdata/test-secrets.json"
)

func Test_YAMLFileWatcher(t *testing.T) {
	initialData := fmt.Sprintf("%s: test-user-1\n%s: test-password-1\n", userName, password)
	overwriteData := fmt.Sprintf("%s: test-user-2\n%s: test-password-2\n", userName, password)
	invalidOverwriteData := "invalid-overwrite-data"

	// write initialData data
	if err := os.WriteFile(yamlFilePath, []byte(initialData), 0644); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second) // wait for the changes to be written

	triggerChan := make(chan struct{}, 1)

	// setup trigger check
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for range triggerChan {
			wg.Done()

			return // return to not react on the error test write
		}
	}()

	enviLoader := envi.NewEnvi()

	// declare a new file watcher with prefix / without setting loader while declaring
	watcher, err := filewatch.NewYAMLFileWatcher(yamlFilePath, enviLoader, filewatch.WithPrefix(prefix), filewatch.WithTriggerChannels(triggerChan))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := watcher.Close(); err != nil {
			t.Error(err)
		}
	})

	var config map[string]string

	// setup error check
	errWG := &sync.WaitGroup{}
	errWG.Add(1)

	// load config files
	if config, err = loadConfig(t, config, enviLoader, errWG, watcher); err != nil {
		t.Fatal(err)
	}

	// assert that the loaded config is the same as initial data
	if config[keyUserName] != "test-user-1" || config[keyPassword] != "test-password-1" {
		t.Error(err)
	}

	t.Run("happy path", func(t *testing.T) {
		// overwrite config files
		if err := os.WriteFile(yamlFilePath, []byte(overwriteData), 0644); err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second) // wait for the changes to be written

		// assert that the loaded config is the same as overwrite data
		if config[keyUserName] != "test-user-2" || config[keyPassword] != "test-password-2" {
			t.Error(err)
		}

		wg.Wait() // wait for triggers to get called
	})

	t.Run("error while file watching", func(t *testing.T) {
		// overwrite config files
		if err := os.WriteFile(yamlFilePath, []byte(invalidOverwriteData), 0644); err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second) // wait for the changes to be written

		errWG.Wait() // wait for error channel to get called
	})

	t.Run("without prefix and loader already set while declaring watcher", func(t *testing.T) {
		// write initialData data
		if err := os.WriteFile(yamlFilePath, []byte(initialData), 0644); err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second) // wait for the changes to be written

		triggerChan := make(chan struct{}, 1)

		// setup trigger check
		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			for range triggerChan {
				wg.Done()

				return
			}
		}()

		// declare a new file watcher without prefix / with setting loader while declaring
		watcher, _ := filewatch.NewYAMLFileWatcher(yamlFilePath, enviLoader, filewatch.WithTriggerChannels(triggerChan))
		t.Cleanup(func() {
			if err := watcher.Close(); err != nil {
				t.Error(err)
			}
		})

		if err := watcher.Start(config); err != nil {
			t.Error(err)
		}

		config = enviLoader.ToMap() // load vars into config map

		// overwrite config files
		if err := os.WriteFile(yamlFilePath, []byte(overwriteData), 0644); err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second) // wait for the changes to be written

		// assert that the loaded config is the same as overwrite data
		if config[userName] != "test-user-2" || config[password] != "test-password-2" {
			t.Error(err)
		}

		wg.Wait() // wait for triggers to get called
	})

	t.Run("empty path", func(t *testing.T) {
		_, err := filewatch.NewYAMLFileWatcher("", nil)
		if err == nil || !errors.Is(err, filewatch.ErrNoPath) {
			t.Error("expected error")
		}
	})

	t.Run("loader not set", func(t *testing.T) {
		_, err := filewatch.NewYAMLFileWatcher("test-path", nil)
		if err == nil || !errors.Is(err, filewatch.ErrLoaderNotSet) {
			t.Error("expected error")
		}
	})

	t.Run("empty prefix", func(t *testing.T) {
		_, err := filewatch.NewYAMLFileWatcher("test-path", enviLoader, filewatch.WithPrefix(""))
		if err == nil || !errors.Is(err, filewatch.ErrEmptyPrefix) {
			t.Error("expected error")
		}
	})

	t.Run("empty prefix", func(t *testing.T) {
		_, err := filewatch.NewYAMLFileWatcher("test-path", enviLoader, filewatch.WithTriggerChannels())
		if err == nil || !errors.Is(err, filewatch.ErrNoTriggers) {
			t.Error("expected error")
		}
	})
}

func Test_JSONFileWatcher(t *testing.T) {
	initialData := fmt.Sprintf(`{"%s": "test-user-1", "%s": "test-password-1"}`, userName, password)
	overwriteData := fmt.Sprintf(`{"%s": "test-user-2", "%s": "test-password-2"}`, userName, password)
	invalidOverwriteData := "invalid-overwrite-data"

	// write initial data
	if err := os.WriteFile(jsonFilePath, []byte(initialData), 0644); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second) // wait for the changes to be written

	triggerChan := make(chan struct{}, 1)

	// setup trigger check
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for range triggerChan {
			wg.Done()

			return
		}
	}()

	enviLoader := envi.NewEnvi()

	// declare a new file watcher with prefix / without setting loader while declaring
	watcher, err := filewatch.NewJSONFileWatcher(jsonFilePath, enviLoader, filewatch.WithPrefix(prefix), filewatch.WithTriggerChannels(triggerChan))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := watcher.Close(); err != nil {
			t.Error(err)
		}
	})

	var config map[string]string

	// setup error check
	errWG := &sync.WaitGroup{}
	errWG.Add(1)

	// load config files
	if config, err = loadConfig(t, config, enviLoader, errWG, watcher); err != nil {
		t.Fatal(err)
	}

	// assert that the loaded config is the same as initial data
	if config[keyUserName] != "test-user-1" || config[keyPassword] != "test-password-1" {
		t.Error(err)
	}

	t.Run("happy path", func(t *testing.T) {
		// overwrite config files
		if err := os.WriteFile(jsonFilePath, []byte(overwriteData), 0644); err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second) // wait for the changes to be written

		// assert that the loaded config is the same as overwrite data
		if config[keyUserName] != "test-user-2" || config[keyPassword] != "test-password-2" {
			t.Error(err)
		}

		wg.Wait() // wait for triggers to get called
	})

	t.Run("error while file watching", func(t *testing.T) {
		// overwrite config files
		if err := os.WriteFile(jsonFilePath, []byte(invalidOverwriteData), 0644); err != nil {
			t.Fatal(err)
		}

		errWG.Wait() // wait for error channel to get called
	})

	t.Run("without prefix", func(t *testing.T) {
		// write initialData data
		if err := os.WriteFile(jsonFilePath, []byte(initialData), 0644); err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second) // wait for the changes to be written

		triggerChan := make(chan struct{}, 1)

		// setup trigger check
		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			for range triggerChan {
				wg.Done()

				return
			}
		}()

		// declare a new file watcher without prefix / with setting loader while declaring
		watcher, _ := filewatch.NewJSONFileWatcher(jsonFilePath, enviLoader, filewatch.WithTriggerChannels(triggerChan))
		t.Cleanup(func() {
			if err := watcher.Close(); err != nil {
				t.Error(err)
			}
		})

		if err := watcher.Start(config); err != nil {
			t.Error(err)
		}

		config = enviLoader.ToMap() // load vars into config map

		// overwrite config files
		if err := os.WriteFile(jsonFilePath, []byte(overwriteData), 0644); err != nil {
			t.Fatal(err)
		}

		time.Sleep(1 * time.Second) // wait for the changes to be written

		// assert that the loaded config is the same as overwrite data
		if config[userName] != "test-user-2" || config[password] != "test-password-2" {
			t.Error(err)
		}

		wg.Wait() // wait for triggers to get called
	})

	t.Run("empty path", func(t *testing.T) {
		_, err := filewatch.NewJSONFileWatcher("", nil)
		if err == nil || !errors.Is(err, filewatch.ErrNoPath) {
			t.Error("expected error")
		}
	})

	t.Run("loader not set", func(t *testing.T) {
		_, err := filewatch.NewJSONFileWatcher("test-path", nil)
		if err == nil || !errors.Is(err, filewatch.ErrLoaderNotSet) {
			t.Error("expected error")
		}
	})

	t.Run("empty prefix", func(t *testing.T) {
		_, err := filewatch.NewJSONFileWatcher("test-path", enviLoader, filewatch.WithPrefix(""))
		if err == nil || !errors.Is(err, filewatch.ErrEmptyPrefix) {
			t.Error("expected error")
		}
	})

	t.Run("empty prefix", func(t *testing.T) {
		_, err := filewatch.NewJSONFileWatcher("test-path", enviLoader, filewatch.WithTriggerChannels())
		if err == nil || !errors.Is(err, filewatch.ErrNoTriggers) {
			t.Error("expected error")
		}
	})
}

func loadConfig(t *testing.T, config map[string]string, enviLoader *envi.Envi, errWG *sync.WaitGroup, watchers ...*filewatch.FileWatcher) (map[string]string, error) {
	t.Helper()

	if len(watchers) > 0 {
		for i := range watchers {
			if err := setupWatcher(t, config, watchers[i], errWG); err != nil {
				return nil, err
			}
		}
	}

	if err := enviLoader.EnsureVars(
		keyUserName,
		keyPassword,
	); err != nil {
		return nil, err
	}

	return enviLoader.ToMap(), nil
}

func setupWatcher(t *testing.T, config map[string]string, w *filewatch.FileWatcher, wg *sync.WaitGroup) error {
	t.Helper()

	if err := w.Start(config); err != nil {
		return err
	}

	go func() {
		errChan := w.ErrChan()

		for range errChan {
			wg.Done()

			return
		}
	}()

	return nil
}
