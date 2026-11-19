package main

import "testing"

func TestAPIPort(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default when empty", input: "", want: "8080"},
		{name: "default when whitespace", input: "  ", want: "8080"},
		{name: "trim and normalize", input: " 08080 ", want: "8080"},
		{name: "lowest valid port", input: "1", want: "1"},
		{name: "highest valid port", input: "65535", want: "65535"},
		{name: "reject non-numeric", input: "http", wantErr: true},
		{name: "reject zero", input: "0", wantErr: true},
		{name: "reject negative", input: "-1", wantErr: true},
		{name: "reject too high", input: "65536", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := apiPort(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("apiPort(%q) error = %v, wantErr %t", test.input, err, test.wantErr)
			}
			if err == nil && got != test.want {
				t.Errorf("apiPort(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
