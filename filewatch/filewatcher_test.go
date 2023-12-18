package filewatch_test

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Clarilab/envi/filewatch/v2"
	"github.com/Clarilab/envi/v2"
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
	orignialData := fmt.Sprintf("%s: test-user-1\n%s: test-password-1\n", userName, password)

	// write original data
	if err := os.WriteFile(yamlFilePath, []byte(orignialData), 0644); err != nil {
		t.Fatal(err)
	}

	triggerChan := make(chan struct{}, 1)

	// setup trigger check
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for range triggerChan {
			wg.Done()
		}
	}()

	// create a new file watcher
	watcher := filewatch.NewYAMLFileWatcher(yamlFilePath, prefix, triggerChan)
	t.Cleanup(func() {
		if err := watcher.Close(); err != nil {
			t.Error(err)
		}
	})

	var config map[string]string
	var err error

	// setup error check
	errWG := &sync.WaitGroup{}
	errWG.Add(1)

	// load config files
	if config, err = loadConfig(t, config, errWG, watcher); err != nil {
		t.Fatal(err)
	}

	// assert that the loaded config is the same as original data
	if config[keyUserName] != "test-user-1" || config[keyPassword] != "test-password-1" {
		t.Error(err)
	}

	t.Run("happy path", func(t *testing.T) {
		// overwrite config files
		overwriteData := fmt.Sprintf("%s: test-user-2\n%s: test-password-2\n", userName, password)

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

	t.Run("error", func(t *testing.T) {
		// overwrite config files
		invalidOverwriteData := "invalid-overwrite-data"

		if err := os.WriteFile(yamlFilePath, []byte(invalidOverwriteData), 0644); err != nil {
			t.Fatal(err)
		}

		errWG.Wait() // wait for error channel to get called
	})
}

func Test_JSONFileWatcher(t *testing.T) {
	originalData := fmt.Sprintf(`{"%s": "test-user-1", "%s": "test-password-1"}`, userName, password)

	// write original data
	if err := os.WriteFile(jsonFilePath, []byte(originalData), 0644); err != nil {
		t.Fatal(err)
	}

	triggerChan := make(chan struct{}, 1)

	// setup trigger check
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for range triggerChan {
			wg.Done()
		}
	}()

	// create a new file watcher
	watcher := filewatch.NewJSONFileWatcher(jsonFilePath, prefix, triggerChan)
	t.Cleanup(func() {
		if err := watcher.Close(); err != nil {
			t.Error(err)
		}
	})

	var config map[string]string
	var err error

	// setup error check
	errWG := &sync.WaitGroup{}
	errWG.Add(1)

	// load config files
	if config, err = loadConfig(t, config, errWG, watcher); err != nil {
		t.Fatal(err)
	}

	// assert that the loaded config is the same as original data
	if config[keyUserName] != "test-user-1" || config[keyPassword] != "test-password-1" {
		t.Error(err)
	}

	t.Run("happy path", func(t *testing.T) {
		overwriteData := fmt.Sprintf(`{"%s": "test-user-2", "%s": "test-password-2"}`, userName, password)

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

	t.Run("error", func(t *testing.T) {
		invalidOverwriteData := "invalid-overwrite-data"

		// overwrite config files
		if err := os.WriteFile(jsonFilePath, []byte(invalidOverwriteData), 0644); err != nil {
			t.Fatal(err)
		}

		errWG.Wait() // wait for error channel to get called
	})
}

func loadConfig(t *testing.T, config map[string]string, errWG *sync.WaitGroup, watchers ...*filewatch.FileWatcher) (map[string]string, error) {
	t.Helper()

	enviLoader := envi.NewEnvi()

	if len(watchers) > 0 {
		for i := range watchers {
			if err := setupWatcher(t, config, watchers[i], enviLoader, errWG); err != nil {
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

func setupWatcher(t *testing.T, config map[string]string, w *filewatch.FileWatcher, loader filewatch.Loader, wg *sync.WaitGroup) error {
	t.Helper()

	w.SetLoader(loader)

	if err := w.Start(config); err != nil {
		return err
	}

	go func() {
		errChan := w.ErrChan()

		for range errChan {
			wg.Done()
		}
	}()

	return nil
}
