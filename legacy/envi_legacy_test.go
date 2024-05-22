package envi_legacy_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Clarilab/envi/v2"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func Test_FromMap(t *testing.T) {
	payload := make(map[string]string)
	payload["EDITOR"] = "vim"
	payload["PAGER"] = "less"

	e := envi.NewEnvi()
	e.FromMap(payload)

	assert.Len(t, e.ToMap(), 2)
}

func Test_LoadEnv(t *testing.T) {
	e := envi.NewEnvi()
	e.LoadEnv("EDITOR", "PAGER", "HOME")

	assert.Len(t, e.ToMap(), 3)
}

func Test_LoadJSONFromFile(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadJSONFiles()

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 0)
	})

	t.Run("a valid json file", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadJSONFiles("testdata/valid1.json")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 3)
	})

	t.Run("2 valid json files", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadJSONFiles("testdata/valid1.json", "testdata/valid2.json")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 4)
	})

	t.Run("an invalid json file", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadJSONFiles("testdata/invalid.json")

		assert.Error(t, err)
	})

	t.Run("a missing file", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadJSONFiles("testdata/idontexist.json")

		assert.Error(t, err)
	})
}

func Test_LoadYAMLFomFile(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadYAMLFiles()

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 0)
	})

	t.Run("a valid yaml file", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadYAMLFiles("testdata/valid1.yaml")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 3)
	})

	t.Run("2 valid yaml files", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadYAMLFiles("testdata/valid1.yaml", "testdata/valid2.yaml")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 4)
	})

	t.Run("an invalid yaml file", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadYAMLFiles("testdata/invalid.yaml")

		assert.Error(t, err)
	})

	t.Run("a missing file", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadYAMLFiles("testdata/idontexist.yaml")

		assert.Error(t, err)
	})
}

func Test_EnsureVars(t *testing.T) {
	t.Run("all ensured vars are present", func(t *testing.T) {
		payload := make(map[string]string)
		payload["EDITOR"] = "vim"
		payload["PAGER"] = "less"

		e := envi.NewEnvi()
		e.FromMap(payload)

		err := e.EnsureVars("EDITOR", "PAGER")

		assert.NoError(t, err)
	})

	t.Run("one ensured var is missing", func(t *testing.T) {
		payload := make(map[string]string)
		payload["EDITOR"] = "vim"
		payload["PAGER"] = "less"

		e := envi.NewEnvi()
		e.FromMap(payload)

		err := e.EnsureVars("EDITOR", "PAGER", "HOME")

		assert.Error(t, err)
	})

	t.Run("all ensured vars are missing", func(t *testing.T) {
		payload := make(map[string]string)
		payload["EDITOR"] = "vim"
		payload["PAGER"] = "less"

		e := envi.NewEnvi()
		e.FromMap(payload)

		err := e.EnsureVars("HOME", "MAIL", "URL")

		assert.Error(t, err)
	})
}

func Test_ToEnv(t *testing.T) {
	payload := make(map[string]string)
	payload["SCHURZLPURZ"] = "yes, indeed"

	e := envi.NewEnvi()
	e.FromMap(payload)

	e.ToEnv()

	assert.Equal(t, "yes, indeed", os.Getenv("SCHURZLPURZ"))
}

func Test_ToMap(t *testing.T) {
	payload := make(map[string]string)
	payload["EDITOR"] = "vim"
	payload["PAGER"] = "less"

	e := envi.NewEnvi()
	e.FromMap(payload)

	vars := e.ToMap()

	assert.Len(t, vars, 2)
}

func Test_LoadFile(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadFile("FILE", "")

		assert.Error(t, err)
		assert.Len(t, e.ToMap(), 0)
	})

	t.Run("file with string content", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadFile("FILE", filepath.Join("testdata/valid.txt"))

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 1)
		assert.Equal(t, "valid string", e.ToMap()["FILE"])
	})
}

func Test_LoadYAMLFilesFromEnvPaths(t *testing.T) {
	const (
		filePath  string = "FILE_PATH"
		filePath2 string = "FILE_PATH2"
		shell     string = "SHELL"
		pager     string = "PAGER"
		calc      string = "CALC"
		mail      string = "MAIL"
	)

	t.Run("1 file", func(t *testing.T) {
		expectedMap := map[string]string{
			"SHELL": "csh",
			"PAGER": "more",
			"CALC":  "bc",
		}

		err := os.Setenv(filePath, "testdata/valid1.yaml")
		assert.NoError(t, err)

		e := envi.NewEnvi()
		err = e.LoadYAMLFilesFromEnvPaths(filePath)
		assert.NoError(t, err)

		err = e.EnsureVars(shell, pager, calc)
		assert.NoError(t, err)

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 3)
		assert.Equal(t, expectedMap, e.ToMap())

		err = os.Unsetenv(filePath)
		assert.NoError(t, err)
	})

	t.Run("multiple files", func(t *testing.T) {
		expectedMap := map[string]string{
			"SHELL": "bash",
			"PAGER": "more",
			"CALC":  "bc",
			"MAIL":  "mutt",
		}

		err := os.Setenv(filePath, "testdata/valid1.yaml")
		assert.NoError(t, err)

		err = os.Setenv(filePath2, "testdata/valid2.yaml")
		assert.NoError(t, err)

		e := envi.NewEnvi()
		err = e.LoadYAMLFilesFromEnvPaths(filePath, filePath2)
		assert.NoError(t, err)

		err = e.EnsureVars(shell, pager, calc)
		assert.NoError(t, err)

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 4)
		assert.Equal(t, expectedMap, e.ToMap())

		err = os.Unsetenv(filePath)
		assert.NoError(t, err)

		err = os.Unsetenv(filePath2)
		assert.NoError(t, err)
	})

	t.Run("no file", func(t *testing.T) {
		err := os.Setenv(filePath, "/file/does/not/exist")
		assert.NoError(t, err)

		e := envi.NewEnvi()
		err = e.LoadYAMLFilesFromEnvPaths(filePath)
		assert.Error(t, err)

		err = os.Unsetenv(filePath)
		assert.NoError(t, err)
	})

	t.Run("env not found", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadYAMLFilesFromEnvPaths(filePath)
		assert.Error(t, err)
	})
}

func Test_LoadJSONFilesFromEnvPaths(t *testing.T) {
	const (
		filePath  string = "FILE_PATH"
		filePath2 string = "FILE_PATH2"
		url       string = "URL"
		editor    string = "EDITOR"
		home      string = "HOME"
		pager     string = "PAGER"
	)

	t.Run("1 file", func(t *testing.T) {
		expectedMap := map[string]string{
			"URL":    "http://foobar.de",
			"EDITOR": "emacs",
			"HOME":   "/home/user",
		}

		err := os.Setenv(filePath, "testdata/valid1.json")
		assert.NoError(t, err)

		e := envi.NewEnvi()
		err = e.LoadJSONFilesFromEnvPaths(filePath)
		assert.NoError(t, err)

		err = e.EnsureVars(url, editor, home)
		assert.NoError(t, err)

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 3)
		assert.Equal(t, expectedMap, e.ToMap())

		err = os.Unsetenv(filePath)
		assert.NoError(t, err)
	})

	t.Run("multiple files", func(t *testing.T) {
		expectedMap := map[string]string{
			"URL":    "http://foobar.de",
			"EDITOR": "vim",
			"HOME":   "/home/user",
			"PAGER":  "less",
		}

		err := os.Setenv(filePath, "testdata/valid1.json")
		assert.NoError(t, err)

		err = os.Setenv(filePath2, "testdata/valid2.json")
		assert.NoError(t, err)

		e := envi.NewEnvi()
		err = e.LoadJSONFilesFromEnvPaths(filePath, filePath2)
		assert.NoError(t, err)

		err = e.EnsureVars(url, editor, home, pager)
		assert.NoError(t, err)

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 4)
		assert.Equal(t, expectedMap, e.ToMap())

		err = os.Unsetenv(filePath)
		assert.NoError(t, err)

		err = os.Unsetenv(filePath2)
		assert.NoError(t, err)
	})

	t.Run("no file", func(t *testing.T) {
		err := os.Setenv(filePath, "/file/does/not/exist")
		assert.NoError(t, err)

		e := envi.NewEnvi()
		err = e.LoadJSONFilesFromEnvPaths(filePath)
		assert.Error(t, err)

		err = os.Unsetenv(filePath)
		assert.NoError(t, err)
	})

	t.Run("env not found", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadJSONFilesFromEnvPaths(filePath)
		assert.Error(t, err)
	})
}

func Test_LoadFileFromEnvPath(t *testing.T) {
	const (
		filePath string = "FILE_PATH"
		key      string = "KEY"
	)

	t.Run("1 file", func(t *testing.T) {
		expectedMap := map[string]string{
			"KEY": "valid string",
		}

		err := os.Setenv(filePath, "testdata/valid.txt")
		assert.NoError(t, err)

		e := envi.NewEnvi()
		err = e.LoadFileFromEnvPath(key, filePath)
		assert.NoError(t, err)

		err = e.EnsureVars(key)
		assert.NoError(t, err)

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 1)
		assert.Equal(t, expectedMap, e.ToMap())

		err = os.Unsetenv(filePath)
		assert.NoError(t, err)
	})

	t.Run("no file", func(t *testing.T) {
		err := os.Setenv(filePath, "/file/does/not/exist")
		assert.NoError(t, err)

		e := envi.NewEnvi()
		err = e.LoadFileFromEnvPath(key, filePath)
		assert.Error(t, err)

		err = os.Unsetenv(filePath)
		assert.NoError(t, err)
	})

	t.Run("env not found", func(t *testing.T) {
		e := envi.NewEnvi()
		err := e.LoadFileFromEnvPath(key, filePath)
		assert.Error(t, err)
	})
}

func Test_LoadAndWatchFile(t *testing.T) {
	const (
		fooKey        = "foo"
		quoKey        = "quo"
		initialFooVal = "bar"
		initialQuoVal = "qux"
		changedQuoVal = "baz"
	)

	testConfig := struct {
		Foo string `json:"foo"`
		Quo string `json:"quo"`
	}{}

	callbackProof := struct {
		wasCalled bool
	}{}

	watcherCallback := func() error {
		callbackProof.wasCalled = true
		return nil
	}

	t.Run("yaml file", func(t *testing.T) {
		const configFilePath = "testdata/watchme.yaml"

		testConfig.Foo = initialFooVal
		testConfig.Quo = initialQuoVal
		callbackProof.wasCalled = false

		// write a yaml file to disk
		blob, err := yaml.Marshal(testConfig)
		assert.NoError(t, err)

		err = os.WriteFile(configFilePath, blob, 0644)
		assert.NoError(t, err)

		// load and watch the yaml file
		e := envi.NewEnvi()

		err, closeFunc, _ := e.LoadAndWatchYAMLFile(configFilePath, watcherCallback)
		assert.NoError(t, err)

		t.Cleanup(func() {
			err := closeFunc()
			if err != nil {
				t.Logf("Failed to close watcher: %v", err)
			}

			os.Remove(configFilePath)
		})

		assert.NoError(t, err)

		err = e.EnsureVars(fooKey, quoKey)
		assert.NoError(t, err)

		// change the file
		testConfig.Quo = changedQuoVal

		newBlob, err := yaml.Marshal(testConfig)
		assert.NoError(t, err)

		err = os.WriteFile(configFilePath, newBlob, 0644)
		assert.NoError(t, err)

		// Wait for changes to take effect (without this, the test fails)
		time.Sleep(100 * time.Millisecond)

		// assert the change is noticed and reflected
		confMap := e.ToMap()

		newQuo, ok := confMap[quoKey]
		assert.True(t, ok)
		assert.Equal(t, changedQuoVal, newQuo)

		// assert the callback has been executed
		assert.True(t, callbackProof.wasCalled)
	})

	t.Run("json file", func(t *testing.T) {
		const configFilePath = "testdata/watchme.json"

		testConfig.Foo = initialFooVal
		testConfig.Quo = initialQuoVal
		callbackProof.wasCalled = false

		// write a json file to disk
		blob, err := json.Marshal(testConfig)
		assert.NoError(t, err)

		err = os.WriteFile(configFilePath, blob, 0644)
		assert.NoError(t, err)

		// load and watch the json file
		e := envi.NewEnvi()

		err, closeFunc, _ := e.LoadAndWatchJSONFile(configFilePath, watcherCallback)
		assert.NoError(t, err)

		t.Cleanup(func() {
			err := closeFunc()
			if err != nil {
				t.Logf("Failed to close watcher: %v", err)
			}

			os.Remove(configFilePath)
		})

		assert.NoError(t, err)

		err = e.EnsureVars(fooKey, quoKey)
		assert.NoError(t, err)

		// change the file
		testConfig.Quo = changedQuoVal

		newBlob, err := json.Marshal(testConfig)
		assert.NoError(t, err)

		err = os.WriteFile(configFilePath, newBlob, 0644)
		assert.NoError(t, err)

		// Wait for changes to take effect (without this, the test fails)
		time.Sleep(100 * time.Millisecond)

		// assert the change is noticed and reflected
		confMap := e.ToMap()

		newQuo, ok := confMap[quoKey]
		assert.True(t, ok)
		assert.Equal(t, changedQuoVal, newQuo)

		// assert the callback has been executed
		assert.True(t, callbackProof.wasCalled)
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		const configFilePath = "I/do/not/exist"

		e := envi.NewEnvi()

		err, _, _ := e.LoadAndWatchJSONFile(configFilePath, nil)
		assert.Error(t, err)
	})

	t.Run("file removed during the watching sends error over channel", func(t *testing.T) {
		const configFilePath = "testdata/watchmegetremoved.json"

		testConfig.Foo = initialFooVal
		testConfig.Quo = initialQuoVal
		callbackProof.wasCalled = false

		blob, err := json.Marshal(testConfig)
		assert.NoError(t, err)

		err = os.WriteFile(configFilePath, blob, 0644)
		assert.NoError(t, err)

		// load and watch the json file
		e := envi.NewEnvi()

		err, closeFunc, watchErrChan := e.LoadAndWatchJSONFile(configFilePath, watcherCallback)
		assert.NoError(t, err)

		t.Cleanup(func() {
			err := closeFunc()
			if err != nil {
				t.Logf("Failed to close watcher: %v", err)
			}
		})

		var watchErrors []error

		go func() {
			for {
				err, ok := <-watchErrChan
				if !ok {
					return
				}

				watchErrors = append(watchErrors, err)
			}
		}()

		time.Sleep(100 * time.Millisecond)

		os.Remove(configFilePath)

		time.Sleep(100 * time.Millisecond)

		assert.NotEmpty(t, watchErrors)

		assert.False(t, callbackProof.wasCalled)
	})

	t.Run("error during callback execution gets sent over channel", func(t *testing.T) {
		const configFilePath = "testdata/watchme.json"

		testConfig.Foo = initialFooVal
		testConfig.Quo = initialQuoVal
		callbackProof.wasCalled = false

		blob, err := json.Marshal(testConfig)
		assert.NoError(t, err)

		err = os.WriteFile(configFilePath, blob, 0644)
		assert.NoError(t, err)

		callbackThatTrowsError := func() error {
			return errors.New("Oh snap!")
		}

		// load and watch the json file
		e := envi.NewEnvi()

		err, closeFunc, watchErrChan := e.LoadAndWatchJSONFile(configFilePath, callbackThatTrowsError)
		assert.NoError(t, err)

		t.Cleanup(func() {
			err := closeFunc()
			if err != nil {
				t.Logf("Failed to close watcher: %v", err)
			}

			os.Remove(configFilePath)
		})

		var watchErrors []error

		go func() {
			for {
				err, ok := <-watchErrChan
				if !ok {
					return
				}

				watchErrors = append(watchErrors, err)
			}
		}()

		// change the file
		testConfig.Quo = changedQuoVal

		newBlob, err := json.Marshal(testConfig)
		assert.NoError(t, err)

		err = os.WriteFile(configFilePath, newBlob, 0644)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		assert.NotEmpty(t, watchErrors)
	})

}
