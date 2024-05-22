package envi_test

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/Clarilab/envi/v2"
	"github.com/davecgh/go-spew/spew"
)

type MightyConfig struct {
	WaitGroup *sync.WaitGroup
	Name      string   `yaml:"PETER"`
	Tenants   []string `yaml:"TENANTS"`
}

type Env struct {
	MightyConfig MightyConfig `env:"ENVI_TEST" watch:"true"`
}

func (m MightyConfig) Notify() {
	m.WaitGroup.Done()
}

func Test_Basic(t *testing.T) {
	t.Setenv("ENVI_TEST", "/Users/maxbreida/dev/github.com/Clarilab/envi/test.yaml")

	env := Env{
		MightyConfig: MightyConfig{
			WaitGroup: new(sync.WaitGroup),
		},
	}

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

	err := envi.GetEnvs(&env)
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(env)

	env.MightyConfig.WaitGroup.Add(2)

	if err := os.WriteFile(
		"test.yaml",
		[]byte(fmt.Sprintf("%s: %s", "PETER", "PANUS")),
		0o664,
	); err != nil {
		t.Fatal(err)
	}

	env.MightyConfig.WaitGroup.Wait()

	if env.MightyConfig.Name != "PANUS" {
		t.Fatal("expected PANUS")
	}

	spew.Dump(env)
}
