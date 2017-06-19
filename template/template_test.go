package template

import "testing"

func TestString(t *testing.T) {
	type args struct {
		s string
		d D
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"base",
			args{"{{.workdir}}/my/path", D{"workdir": "/home/me"}},
			"/home/me/my/path",
			false},
		{"base",
			args{"{{.pwd}}/my/path", D{"workdir": "/home/me"}},
			"",
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := String(tt.args.s, tt.args.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
