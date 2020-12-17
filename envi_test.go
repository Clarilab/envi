package envi

import "testing"

func Test_LoadEnvVars_Empty(t *testing.T) {
	loadedVars, err := LoadEnvVars([]string{})
	if err != nil {
		t.Error(err)
	}

	if len(loadedVars) != 0 {
		t.Fail()
	}
}

func Test_LoadEnvVars_Nil(t *testing.T) {
	loadedVars, err := LoadEnvVars(nil)
	if err != nil {
		t.Error(err)
	}

	if len(loadedVars) != 0 {
		t.Fail()
	}
}

func Test_LoadEnvVars_All(t *testing.T) {
	required := []string{"EDITOR", "HOME"}

	loadedVars, err := LoadEnvVars(required)
	if err != nil {
		t.Error(err)
	}

	if loadedVars["EDITOR"] == "" {
		t.Error("EDITOR was not set, but expected.")
	}

	if loadedVars["HOME"] == "" {
		t.Error("HOME was not set, but expected.")
	}
}

func Test_LoadEnvVars_Missing(t *testing.T) {
	required := []string{"EDITOR", "HOME", "SCHNURZLPUTZ"}

	loadedVars, err := LoadEnvVars(required)
	if loadedVars != nil && loadedVars["SCHNURZLPUTZ"] != "" {
		t.Error("SCHNURZLPUTZ was expected to be not set as Environment Variable.")
	}

	if err == nil {
		t.Error("No Error was given while missing an required Environment Variable.")
	}
}

func Test_LoadEnvVarsWithOptional_Empty(t *testing.T) {

	loadedVars, err := LoadEnvVarsWithOptional([]string{}, []string{})
	if err != nil {
		t.Error(err)
	}

	if len(loadedVars) != 0 {
		t.Fail()
	}
}

func Test_LoadEnvVarsWithOptional_OnlyOptional(t *testing.T) {
	optional := []string{"SCHNURZLPUTZ"}

	loadedVars, err := LoadEnvVarsWithOptional([]string{}, optional)
	if err != nil {
		t.Error(err)
	}

	if loadedVars["SCHNURZLPUTZ"] != "" {
		t.Error("Didn't expect SCHNURZLPUTZ to be set.")
	}
}

func Test_LoadEnvVarsWithOptional_OnlyRequired(t *testing.T) {
	required := []string{"EDITOR", "HOME"}

	loadedVars, err := LoadEnvVarsWithOptional(required, []string{})
	if err != nil {
		t.Error(err)
	}

	if loadedVars["EDITOR"] == "" {
		t.Error("EDITOR was not set, but expected.")
	}

	if loadedVars["HOME"] == "" {
		t.Error("HOME was not set, but expected.")
	}
}

func Test_LoadEnvVarsWithOptional_MissingRequired(t *testing.T) {
	required := []string{"EDITOR", "HOME", "SCHNURZLPUTZ"}

	loadedVars, err := LoadEnvVarsWithOptional(required, []string{})
	if loadedVars != nil && loadedVars["SCHNURZLPUTZ"] != "" {
		t.Error("SCHNURZLPUTZ was expected to be not set as Environment Variable.")
	}

	if err == nil {
		t.Error("No Error was given while missing an required Environment Variable.")
	}
}
