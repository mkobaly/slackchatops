package slackchatops

import "testing"

func TestParseArgs(t *testing.T) {
	action := Action{Name: "Foo", Params: []string{"id", "name"}, Args: []string{"-c", "{0}", "{0} | {1}"}}
	result := action.ParseArgs([]string{"232", "bar"})

	if len(result) != 3 {
		t.Error("Args length is not 3")
	}
	if result[1] != "232" {
		t.Error("Merging of args failed")
	}
}

func TestValidateArgs(t *testing.T) {
	action := Action{Name: "Foo", Params: []string{"id", "name"}, Args: []string{"-c", "{0}", "{0} | {1}"}}
	result := action.ValidateArgs()
	if result != nil {
		t.Error(result)
	}
}

func TestValidateArgsWithNotEnoughParams(t *testing.T) {
	action := Action{Name: "Foo", Params: []string{"id"}, Args: []string{"-c", "{0}", "{0} | {1}"}}
	result := action.ValidateArgs()
	if result != nil {
		t.Error(result)
	}
}
