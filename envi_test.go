package envi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FromMap(t *testing.T) {
	payload := make(map[string]string)
	payload["EDITOR"] = "vim"
	payload["PAGER"] = "less"

	e := NewEnvi()
	e.FromMap(payload)

	assert.Len(t, e.ToMap(), 2)
}

func Test_LoadEnv(t *testing.T) {
	e := NewEnvi()
	e.LoadEnv("EDITOR", "PAGER", "HOME")

	assert.Len(t, e.ToMap(), 3)
}

func Test_LoadJSONFromFile(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadJSONFiles()

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 0)
	})

	t.Run("a valid json file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadJSONFiles("testdata/valid1.json")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 3)
	})

	t.Run("2 valid json files", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadJSONFiles("testdata/valid1.json", "testdata/valid2.json")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 4)
	})

	t.Run("an invalid json file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadJSONFiles("testdata/invalid.json")

		assert.Error(t, err)
	})

	t.Run("a missing file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadJSONFiles("testdata/idontexist.json")

		assert.Error(t, err)
	})
}

func Test_LoadYAMLFomFile(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadYAMLFiles()

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 0)
	})

	t.Run("a valid yaml file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadYAMLFiles("testdata/valid1.yaml")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 3)
	})

	t.Run("2 valid yaml files", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadYAMLFiles("testdata/valid1.yaml", "testdata/valid2.yaml")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 4)
	})

	t.Run("an invalid yaml file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadYAMLFiles("testdata/invalid.yaml")

		assert.Error(t, err)
	})

	t.Run("a missing file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadYAMLFiles("testdata/idontexist.yaml")

		assert.Error(t, err)
	})
}

func Test_EnsureVars(t *testing.T) {
	t.Run("all ensured vars are present", func(t *testing.T) {
		payload := make(map[string]string)
		payload["EDITOR"] = "vim"
		payload["PAGER"] = "less"

		e := NewEnvi()
		e.FromMap(payload)

		err := e.EnsureVars("EDITOR", "PAGER")

		assert.NoError(t, err)
	})

	t.Run("one ensured var is missing", func(t *testing.T) {
		payload := make(map[string]string)
		payload["EDITOR"] = "vim"
		payload["PAGER"] = "less"

		e := NewEnvi()
		e.FromMap(payload)

		err := e.EnsureVars("EDITOR", "PAGER", "HOME")

		assert.Error(t, err)
	})

	t.Run("all ensured vars are missing", func(t *testing.T) {
		payload := make(map[string]string)
		payload["EDITOR"] = "vim"
		payload["PAGER"] = "less"

		e := NewEnvi()
		e.FromMap(payload)

		err := e.EnsureVars("HOME", "MAIL", "URL")

		assert.Error(t, err)
	})
}

func Test_ToEnv(t *testing.T) {
	payload := make(map[string]string)
	payload["SCHURZLPURZ"] = "yes, indeed"

	e := NewEnvi()
	e.FromMap(payload)

	e.ToEnv()

	assert.Equal(t, "yes, indeed", os.Getenv("SCHURZLPURZ"))
}

func Test_ToMap(t *testing.T) {
	payload := make(map[string]string)
	payload["EDITOR"] = "vim"
	payload["PAGER"] = "less"

	e := NewEnvi()
	e.FromMap(payload)

	vars := e.ToMap()

	assert.Len(t, vars, 2)
}

func Test_LoadFile(t *testing.T) {
	t.Run("no file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadFile("FILE", "")

		assert.Error(t, err)
		assert.Len(t, e.ToMap(), 0)
	})

	t.Run("file with string content", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadFile("FILE", filepath.Join("testdata/valid.txt"))

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 1)
		assert.Equal(t, "valid string", e.ToMap()["FILE"])
	})
}