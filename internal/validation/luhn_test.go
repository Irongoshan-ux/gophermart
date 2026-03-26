package validation

import "testing"

func TestLuhnValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid 1", "12345678903", true},
		{"valid 2", "9278923470", true},
		{"valid 3", "346436439", true},
		{"invalid 1", "12345678902", false},
		{"empty", "", false},
		{"non-digit", "12a34", false},
		{"single digit", "7", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LuhnValid(tt.input); got != tt.want {
				t.Errorf("LuhnValid(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
