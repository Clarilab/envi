package envi_test

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/Clarilab/envi/v2"
)

// !!! Attention: The tests in this file are not meant to be run in parallel because of the t.Setenv usage !!!

func Test_DefaultTag(t *testing.T) {
	type Config struct {
		Peter       string `default:"PAN"`
		Environment string `env:"ENVIRONMENT"`
		ServiceName string `env:"SERVICE_NAME" default:"envi-test"`
	}

	testCases := map[string]struct {
		config         Config
		expectedConfig Config
		envvars        map[string]string
		expectedErr    error
	}{
		"empty config with no envvars set": {
			config: Config{},
			expectedConfig: Config{
				Peter:       "PAN",
				ServiceName: "envi-test",
			},
			envvars:     nil,
			expectedErr: nil,
		},
		"empty config with envvars set overwrites defaults": {
			config: Config{},
			expectedConfig: Config{
				Peter:       "PAN",
				Environment: "dev",
				ServiceName: "my-service",
			},
			envvars: map[string]string{
				"ENVIRONMENT":  "dev",
				"SERVICE_NAME": "my-service",
			},
			expectedErr: nil,
		},
		"pre filled config gets overwritten": {
			config: Config{
				Peter:       "Panus",
				Environment: "prod",
				ServiceName: "your-service",
			},
			expectedConfig: Config{
				Peter:       "PAN",
				Environment: "dev",
				ServiceName: "my-service",
			},
			envvars: map[string]string{
				"ENVIRONMENT":  "dev",
				"SERVICE_NAME": "my-service",
			},
			expectedErr: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			for k, v := range tc.envvars {
				t.Setenv(k, v)
			}

			e := envi.New()

			err := e.Load(&tc.config)
			switch {
			case err != nil && tc.expectedErr == nil:
				t.Errorf("expected no error but got %v", err)
			case err == nil && tc.expectedErr != nil:
				t.Errorf("expected error %v but got nil", tc.expectedErr)
			case err != nil && tc.expectedErr != nil:
				if errors.Unwrap(err).Error() != tc.expectedErr.Error() {
					t.Errorf("expected error %v but got %v", tc.expectedErr, err)
				}
			case err == nil && tc.expectedErr == nil:
				if tc.config != tc.expectedConfig {
					t.Errorf("expected config %+v but got %+v", tc.expectedConfig, tc.config)
				}
			}
		})
	}
}

func Test_RequiredTag(t *testing.T) {
	type Config struct {
		Peter       string `default:"PAN" required:"true"`
		Environment string `env:"ENVIRONMENT" required:"true"`
		ServiceName string `env:"SERVICE_NAME"`
	}

	testCases := map[string]struct {
		config         Config
		expectedConfig Config
		envvars        map[string]string
		expectedErr    error
	}{
		"required field missing returns error": {
			config:         Config{},
			expectedConfig: Config{},
			envvars:        nil,
			expectedErr: &envi.ValidationError{
				[]error{&envi.FieldRequiredError{
					FieldName: "Environment",
				}},
			},
		},
		"required fields all present is passes validation": {
			config: Config{},
			expectedConfig: Config{
				Peter:       "PAN",
				Environment: "dev",
			},
			envvars: map[string]string{
				"ENVIRONMENT": "dev",
			},
			expectedErr: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			for k, v := range tc.envvars {
				t.Setenv(k, v)
			}

			e := envi.New()

			err := e.Load(&tc.config)
			switch {
			case err != nil && tc.expectedErr == nil:
				t.Errorf("expected no error but got %v", err)
			case err == nil && tc.expectedErr != nil:
				t.Errorf("expected error %v but got nil", tc.expectedErr)
			case err != nil && tc.expectedErr != nil:
				if errors.Unwrap(err).Error() != tc.expectedErr.Error() {
					t.Errorf("expected error %v but got %v", tc.expectedErr, err)
				}
			case err == nil && tc.expectedErr == nil:
				if tc.config != tc.expectedConfig {
					t.Errorf("expected config %+v but got %+v", tc.expectedConfig, tc.config)
				}
			}
		})
	}
}

type MightyConfig struct {
	WaitGroup *sync.WaitGroup
	Name      string   `yaml:"PETER" required:"true"`
	Tenants   []string `yaml:"TENANTS"`
}

type Config struct {
	MightyConfig MightyConfig `env:"ENVI_TEST111" watch:"true" default:"./test.yaml"`
	ServiceName  string       `env:"SERVICE_NAME" default:"envi-test"`
}

func (m MightyConfig) OnChange() {
	m.WaitGroup.Done()
}

func (m MightyConfig) OnError(err error) {
	fmt.Println(err)
}

func Test_Filewatcher(t *testing.T) {
	t.Setenv("ENVI_TEST", "./test.yaml")

	config := Config{
		MightyConfig: MightyConfig{
			WaitGroup: new(sync.WaitGroup),
		},
	}

	enviClient := envi.New()

	if err := os.WriteFile(
		"test.yaml",
		[]byte(fmt.Sprintf("%s: %s", "PETER", "PAN")),
		0o664,
	); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Remove("test.yaml"); err != nil {
			t.Fatal(err)
		}
	})

	err := enviClient.Load(&config)
	if err != nil {
		t.Fatal(err)
	}

	config.MightyConfig.WaitGroup.Add(1)

	if err := os.WriteFile(
		"test.yaml",
		[]byte(fmt.Sprintf("%s: %s", "PETER", "PANUS")),
		0o664,
	); err != nil {
		t.Fatal(err)
	}

	config.MightyConfig.WaitGroup.Wait()

	if config.MightyConfig.Name != "PANUS" {
		t.Fatal("expected PANUS")
	}

	err = enviClient.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ParseFiles(t *testing.T) {
	type YAMLFile struct {
		PETER string `yaml:"PETER"`
	}

	type JSONFile struct {
		GUENTHER string `json:"GUENTHER"`
	}

	type TextFile struct {
		Value string
	}

	type Config struct {
		YamlFile YAMLFile `default:"./test.yaml" type:"yaml"`
		JsonFile JSONFile `default:"./test.json" type:"json"`
		TextFile TextFile `default:"./test" type:"text"`
	}

	if err := os.WriteFile(
		"test.yaml",
		[]byte(fmt.Sprintf("%s: %s", "PETER", "PAN")),
		0o664,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		"test.json",
		[]byte(fmt.Sprintf("{\"%s\": \"%s\"}", "GUENTHER", "NETZER")),
		0o664,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		"test",
		[]byte("foobar"),
		0o664,
	); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Remove("test.yaml"); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove("test.json"); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove("test"); err != nil {
			t.Fatal(err)
		}
	})

	var myConfig Config

	enviClient := envi.New()
	err := enviClient.Load(&myConfig)
	if err != nil {
		t.Fatal(err)
	}

	if myConfig.YamlFile.PETER != "PAN" {
		t.Fatal("expected PAN")
	}

	if myConfig.JsonFile.GUENTHER != "NETZER" {
		t.Fatal("expected NETZER")
	}

	if myConfig.TextFile.Value != "foobar" {
		t.Fatal("expected foobar")
	}
}
