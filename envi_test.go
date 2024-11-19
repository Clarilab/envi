package envi_test

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/Clarilab/envi/v3"
)

// !!! Attention: The tests in this file are not meant to be run in parallel because of the t.Setenv usage !!!

func Test_DefaultTag(t *testing.T) {
	type Config struct {
		Peter       string `default:"PAN"`
		Environment string `env:"ENVIRONMENT"`
		ServiceName string `default:"envi-test" env:"SERVICE_NAME"`
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
	callbackCounter *atomic.Int32
	Name            string   `required:"true" yaml:"PETER"`
	Tenants         []string `yaml:"TENANTS"`
}

type Config struct {
	MightyConfig MightyConfig `default:"./test.yaml" env:"ENVI_TEST111" watch:"true"`
	ServiceName  string       `default:"envi-test" env:"SERVICE_NAME"`
}

func (m MightyConfig) OnChange() {
	m.callbackCounter.Add(1)
}

func (m MightyConfig) OnError(err error) {
	fmt.Println(err)
}

func Test_Filewatcher(t *testing.T) {
	t.Setenv("ENVI_TEST", "./test.yaml")

	config := Config{
		MightyConfig: MightyConfig{
			callbackCounter: new(atomic.Int32),
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

	if err := os.WriteFile(
		"test.yaml",
		[]byte(fmt.Sprintf("%s: %s", "PETER", "PANUS")),
		0o664,
	); err != nil {
		t.Fatal(err)
	}

	for config.MightyConfig.callbackCounter.Load() < 1 {
		// wait for the callback
	}

	if config.MightyConfig.Name != "PANUS" {
		t.Fatal("expected PANUS")
	}

	err = enviClient.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ParseFiles(t *testing.T) {
	type JSONFile struct {
		URL    string `json:"URL"`
		Editor string `json:"EDITOR"`
		Home   string `json:"HOME"`
	}

	type YAMLFile struct {
		Shell string `yaml:"SHELL"`
		Pager string `yaml:"PAGER"`
		Calc  string `yaml:"CALC"`
	}

	type TextFile struct {
		Value string
	}

	type Config struct {
		JsonFile JSONFile `default:"./testdata/valid.json" type:"json"`
		YamlFile YAMLFile `default:"./testdata/valid.yaml" type:"yaml"`
		TextFile TextFile `default:"./testdata/valid.txt" type:"text"`
	}

	var myConfig Config

	enviClient := envi.New()
	err := enviClient.Load(&myConfig)
	if err != nil {
		t.Fatal(err)
	}

	expectedConfig := Config{
		JsonFile: JSONFile{
			URL:    "http://foobar.de",
			Editor: "emacs",
			Home:   "/home/user",
		},
		YamlFile: YAMLFile{
			Shell: "csh",
			Pager: "more",
			Calc:  "bc",
		},
		TextFile: TextFile{
			Value: "valid string",
		},
	}

	if myConfig != expectedConfig {
		t.Errorf("expected %+v but got %+v", expectedConfig, myConfig)
	}
}

func Test_UnexportedFields(t *testing.T) {
	type ConfigWithUnexportedField struct {
		unexported  string
		Peter       string `default:"PAN"`
		Environment string `env:"ENVIRONMENT"`
		ServiceName string `default:"envi-test" env:"SERVICE_NAME"`
	}

	testCases := map[string]struct {
		config         ConfigWithUnexportedField
		expectedConfig ConfigWithUnexportedField
		envvars        map[string]string
		expectedErr    error
	}{
		"unexported fields do not require a default or env tag": {
			config: ConfigWithUnexportedField{
				unexported:  "foo",
				Peter:       "",
				Environment: "",
				ServiceName: "",
			},
			expectedConfig: ConfigWithUnexportedField{
				unexported:  "foo",
				Peter:       "PAN",
				Environment: "test",
				ServiceName: "my-service",
			},
			envvars: map[string]string{
				"ENVIRONMENT":  "test",
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
