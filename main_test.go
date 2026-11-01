package main

import (
	"errors"
	"testing"
)

func Test_ParseArgs(t *testing.T) {
	for _, tc := range []struct {
		args    []string
		login   string
		wantErr bool
	}{
		{args: nil, login: ""},
		{args: []string{"octocat"}, login: "octocat"},
		{args: []string{" octocat "}, login: "octocat"},
		{args: []string{"octocat", "hubot"}, wantErr: true},
		{args: []string{"--theme"}, wantErr: true},
		{args: []string{""}, wantErr: true},
	} {
		login, err := parseArgs(tc.args)
		if (err != nil) != tc.wantErr || login != tc.login {
			t.Errorf("parseArgs(%q) = %q, %v; want %q, error %v", tc.args, login, err, tc.login, tc.wantErr)
		}
	}

	for _, help := range []string{"-h", "--help"} {
		if _, err := parseArgs([]string{help}); !errors.Is(err, errHelp) {
			t.Errorf("parseArgs(%q) = %v, want the usage text", help, err)
		}
	}
}
