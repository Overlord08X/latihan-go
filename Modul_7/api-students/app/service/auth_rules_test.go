package service

import "testing"

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "password valid — 8 karakter unik",
			password: "K0piSusu!",
			wantErr:  false,
		},
		{
			name:     "password valid — panjang dengan spesial karakter",
			password: "Pr@ktikum-Backend2024",
			wantErr:  false,
		},
		{
			name:     "password valid — angka dan huruf tidak umum",
			password: "zXq9mN3r",
			wantErr:  false,
		},
		{
			name:     "password terlalu pendek — 7 karakter",
			password: "abc1234",
			wantErr:  true,
		},
		{
			name:     "password terlalu pendek — kosong",
			password: "",
			wantErr:  true,
		},
		{
			name:     "password terlalu umum — password123",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "password terlalu umum — PASSWORD123 (case insensitive)",
			password: "PASSWORD123",
			wantErr:  true,
		},
		{
			name:     "password terlalu umum — 12345678",
			password: "12345678",
			wantErr:  true,
		},
		{
			name:     "password terlalu umum — admin123",
			password: "admin123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordStrength(%q) error = %v, wantErr = %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestValidateRegister(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		password string
		wantErr  bool
		errKey   string
	}{
		{
			name:     "valid semua field",
			username: "sari",
			email:    "sari@example.com",
			password: "K0piSusu!",
			wantErr:  false,
		},
		{
			name:     "username kosong",
			username: "",
			email:    "sari@example.com",
			password: "K0piSusu!",
			wantErr:  true,
			errKey:   "username",
		},
		{
			name:     "email tidak ada @",
			username: "sari",
			email:    "sari-example.com",
			password: "K0piSusu!",
			wantErr:  true,
			errKey:   "email",
		},
		{
			name:     "password terlalu umum",
			username: "sari",
			email:    "sari@example.com",
			password: "password1",
			wantErr:  true,
			errKey:   "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, err := ValidateRegister(tt.username, tt.email, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRegister() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errKey != "" {
				if _, ok := errs[tt.errKey]; !ok {
					t.Errorf("ValidateRegister() errs = %v, harus mengandung key %q", errs, tt.errKey)
				}
			}
		})
	}
}
