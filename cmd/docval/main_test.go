package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

// capture executa run() com os argumentos dados, capturando tudo que é
// escrito em stdout e stderr, e devolve os dois textos junto com o código
// de saída retornado.
func capture(t *testing.T, args []string) (stdout, stderr string, code int) {
	t.Helper()

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	code = run(args, outW, errW)

	outW.Close()
	errW.Close()

	outBytes, _ := io.ReadAll(outR)
	errBytes, _ := io.ReadAll(errR)

	return string(outBytes), string(errBytes), code
}

func TestRunCPFValidar(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
	}{
		{"cpf válido", []string{"cpf", "validar", "111.444.777-35"}, 0, "111.444.777-35: válido\n"},
		{"cpf válido sem formatação", []string{"cpf", "validar", "11144477735"}, 0, "111.444.777-35: válido\n"},
		{"cpf inválido", []string{"cpf", "validar", "111.444.777-36"}, 1, ""},
		{"cpf mal formado", []string{"cpf", "validar", "123"}, 1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, _, code := capture(t, tt.args)
			if code != tt.wantCode {
				t.Errorf("código de saída = %d, quero %d (stdout=%q)", code, tt.wantCode, out)
			}
			if tt.wantOut != "" && out != tt.wantOut {
				t.Errorf("stdout = %q, quero %q", out, tt.wantOut)
			}
		})
	}
}

func TestRunCPFFormatar(t *testing.T) {
	out, _, code := capture(t, []string{"cpf", "formatar", "11144477735"})
	if code != 0 {
		t.Fatalf("código de saída = %d, quero 0", code)
	}
	want := "111.444.777-35\n"
	if out != want {
		t.Errorf("stdout = %q, quero %q", out, want)
	}
}

func TestRunCPFGerar(t *testing.T) {
	out, _, code := capture(t, []string{"cpf", "gerar"})
	if code != 0 {
		t.Fatalf("código de saída = %d, quero 0", code)
	}
	out = strings.TrimSpace(out)

	// O CPF gerado deve validar por meio da própria CLI.
	_, _, validarCode := capture(t, []string{"cpf", "validar", out})
	if validarCode != 0 {
		t.Errorf("CPF gerado %q não passou na validação (código=%d)", out, validarCode)
	}
}

func TestRunSemArgumentos(t *testing.T) {
	_, errOut, code := capture(t, []string{})
	if code != 2 {
		t.Errorf("código de saída = %d, quero 2", code)
	}
	if errOut == "" {
		t.Error("esperava mensagem de uso em stderr")
	}
}

func TestRunComandoDesconhecido(t *testing.T) {
	_, errOut, code := capture(t, []string{"xyz"})
	if code != 2 {
		t.Errorf("código de saída = %d, quero 2", code)
	}
	if !strings.Contains(errOut, "xyz") {
		t.Errorf("stderr = %q, esperava menção ao comando desconhecido", errOut)
	}
}

func TestRunVersion(t *testing.T) {
	out, _, code := capture(t, []string{"version"})
	if code != 0 {
		t.Errorf("código de saída = %d, quero 0", code)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("esperava a versão em stdout")
	}
}
