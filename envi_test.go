package envi_test

import (
	"testing"

	"github.com/Clarilab/envi/v2"
	"github.com/davecgh/go-spew/spew"
)

func Test_Basic(t *testing.T) {
	t.Setenv("ENVI_TEST", "/Users/lks/development/go/playground/mighty.yml")

	type MightyConfig struct {
		Name    string   `yaml:"PETER"`
		Tenants []string `yaml:"TENANTS"`
	}

	type Env struct {
        MightyConfig MightyConfig `env:"ENVI_TEST"`
	}

	env := Env{}

	envi.GetEnvs(&env)

	spew.Dump(env)
}
