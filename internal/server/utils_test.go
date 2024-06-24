package server

import "testing"

func Test_isValidLuhnAlgorithm(t *testing.T) {
	tests := []struct {
		name   string
		number int
		want   bool
	}{
		{
			name:   "some valid debit card number",
			number: 4099013600418229,
			want:   true,
		},
		{
			name:   "some Luhn-valid(618304455) number",
			number: 618304455,
			want:   true,
		},
		{
			name:   "some Luhn-valid(1847803446726) number",
			number: 1847803446726,
			want:   true,
		},
		{
			name:   "some Luhn-valid(1234567897) number",
			number: 1234567897,
			want:   true,
		},
		{
			name:   "some Luhn-invalid number",
			number: 4561261212345464,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidLuhnAlgorithm(tt.number); got != tt.want {
				t.Errorf("isValidLuhnAlgorithm() = %v, want %v", got, tt.want)
			}
		})
	}
}
