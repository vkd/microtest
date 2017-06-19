package main

import "testing"

func TestMockConfig_equalURL(t *testing.T) {
	tests := []struct {
		name      string
		configURL string
		expectURL string
		wantErr   bool
	}{
		// TODO: Add test cases.
		{"base", "/hello", "/hello", false},
		{"base query", "/hello?name=mike&age=12,14", "/hello?age=12,14&name=mike", false},

		{"false query", "/hello?name=mike&age=11", "/hello?age=12&name=mike", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockConfig{
				URL: tt.configURL,
				Out: `{"result": "ok"}`,
			}
			if err := m.equalURL(tt.expectURL); (err != nil) != tt.wantErr {
				t.Errorf("MockConfig.equalURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
