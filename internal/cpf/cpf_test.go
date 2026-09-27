package cpf

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"válido formatado", "111.444.777-35", nil},
		{"válido sem formatação", "11144477735", nil},
		{"válido com espaços", " 111.444.777-35 ", nil},
		{"outro cpf válido", "529.982.247-25", nil},
		{"dígitos verificadores errados", "111.444.777-36", ErrCheckDigits},
		{"todos os dígitos iguais", "111.111.111-11", ErrAllDigitsEqual},
		{"todos zeros", "000.000.000-00", ErrAllDigitsEqual},
		{"curto demais", "123.456.789", ErrInvalidLength},
		{"longo demais", "111.444.777-355", ErrInvalidLength},
		{"vazio", "", ErrInvalidLength},
		{"apenas letras", "abc.def.ghi-jk", ErrInvalidLength},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate(%q) = %v, quero %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	if !IsValid("111.444.777-35") {
		t.Error("esperava que 111.444.777-35 fosse válido")
	}
	if IsValid("111.111.111-11") {
		t.Error("esperava que 111.111.111-11 fosse inválido")
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"11144477735", "111.444.777-35", false},
		{"111.444.777-35", "111.444.777-35", false},
		{"123", "", true},
	}

	for _, tt := range tests {
		got, err := Format(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("Format(%q): esperava erro, obteve nil", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("Format(%q): erro inesperado: %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("Format(%q) = %q, quero %q", tt.input, got, tt.want)
		}
	}
}

func TestStrip(t *testing.T) {
	got := Strip("111.444.777-35")
	want := "11144477735"
	if got != want {
		t.Errorf("Strip() = %q, quero %q", got, want)
	}
}

// TestGenerate garante que Generate sempre produz um CPF formatado e válido,
// rodando várias vezes para reduzir a chance de um teste "com sorte".
func TestGenerate(t *testing.T) {
	for i := 0; i < 200; i++ {
		got := Generate()
		if err := Validate(got); err != nil {
			t.Fatalf("Generate() produziu CPF inválido %q: %v", got, err)
		}
	}
}

func BenchmarkValidate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Validate("111.444.777-35")
	}
}
