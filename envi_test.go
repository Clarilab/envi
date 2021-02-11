package envi

import (
	"os"
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
		err := e.LoadJSONFiles("test/valid1.json")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 3)
	})

	t.Run("2 valid json files", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadJSONFiles("test/valid1.json", "test/valid2.json")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 4)
	})

	t.Run("an invalid json file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadJSONFiles("test/invalid.json")

		assert.Error(t, err)
	})

	t.Run("a missing file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadJSONFiles("test/idontexist.json")

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
		err := e.LoadYAMLFiles("test/valid1.yaml")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 3)
	})

	t.Run("2 valid yaml files", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadYAMLFiles("test/valid1.yaml", "test/valid2.yaml")

		assert.NoError(t, err)
		assert.Len(t, e.ToMap(), 4)
	})

	t.Run("an invalid yaml file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadYAMLFiles("test/invalid.yaml")

		assert.Error(t, err)
	})

	t.Run("a missing file", func(t *testing.T) {
		e := NewEnvi()
		err := e.LoadYAMLFiles("test/idontexist.yaml")

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
