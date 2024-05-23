package envi_test

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/Clarilab/envi/v2"
)

type MightyConfig struct {
	WaitGroup *sync.WaitGroup
	Name      string   `yaml:"PETER" required:"true"`
	Tenants   []string `yaml:"TENANTS"`
	Foo       string   `yaml:"FOO" default:"bar"`
	Int32     int32    `default:"123"`
	Int64     int64    `default:"123456"`
}

type Config struct {
	MightyConfig MightyConfig `env:"ENVI_TEST" watch:"true" default:"/Users/maxbreida/dev/github.com/Clarilab/envi/test.yaml"`
	TextFile     Textfile     `env:"TEXT_FILE" type:"text" default:"/Users/maxbreida/dev/github.com/Clarilab/envi/test.yaml"`
	ServiceName  string       `env:"SERVICE_NAME" default:"envi-test"`
}

type Textfile struct {
	Text  string `default:"blabla"`
	Text2 string `default:"blabla2"`
}

func (m MightyConfig) OnChange() {
	m.WaitGroup.Done()
}

func (m MightyConfig) OnError(err error) {
	fmt.Println(err)
}

func Test_Basic(t *testing.T) {
	t.Setenv("ENVI_TEST", "/Users/maxbreida/dev/github.com/Clarilab/envi/test.yaml")

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

	err := enviClient.GetConfig(&config)
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
