// Comando docval é uma CLI para validar, formatar e gerar CPFs.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/beatrizmenezes/docval/internal/cpf"
)

// version é sobrescrita em tempo de build, por exemplo:
//
//	go build -ldflags "-X main.version=v1.0.0"
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run contém toda a lógica da CLI e devolve o código de saída do processo.
// Recebe os writers de stdout/stderr explicitamente para poder ser testada
// sem depender de os.Stdout/os.Stderr globais.
func run(args []string, stdout, stderr *os.File) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "-v", "--version", "version":
		fmt.Fprintln(stdout, version)
		return 0
	case "-h", "--help", "help":
		printUsage(stdout)
		return 0
	case "cpf":
		return runCPF(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "docval: comando desconhecido %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func printUsage(w *os.File) {
	fmt.Fprintln(w, `docval - validador de documentos (CPF)

Uso:
  docval cpf validar <numero>     valida um CPF
  docval cpf formatar <numero>    formata um CPF como 000.000.000-00
  docval cpf gerar                gera um CPF válido para testes
  docval version                  mostra a versão
  docval help                     mostra esta ajuda

Exemplos:
  docval cpf validar 111.444.777-35
  docval cpf validar 11144477735
  docval cpf formatar 11144477735
  docval cpf gerar`)
}

func runCPF(args []string, stdout, stderr *os.File) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "docval cpf: informe uma subação: validar, formatar ou gerar")
		return 2
	}

	action := args[0]

	switch action {
	case "gerar":
		fmt.Fprintln(stdout, cpf.Generate())
		return 0

	case "validar":
		fs := flag.NewFlagSet("docval cpf validar", flag.ContinueOnError)
		fs.SetOutput(stderr)
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if fs.NArg() != 1 {
			fmt.Fprintln(stderr, "docval cpf validar: informe exatamente um número de CPF")
			return 2
		}

		numero := fs.Arg(0)
		if err := cpf.Validate(numero); err != nil {
			fmt.Fprintf(stdout, "%s: inválido (%s)\n", numero, err)
			return 1
		}
		formatado, _ := cpf.Format(numero)
		fmt.Fprintf(stdout, "%s: válido\n", formatado)
		return 0

	case "formatar":
		fs := flag.NewFlagSet("docval cpf formatar", flag.ContinueOnError)
		fs.SetOutput(stderr)
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if fs.NArg() != 1 {
			fmt.Fprintln(stderr, "docval cpf formatar: informe exatamente um número de CPF")
			return 2
		}

		formatado, err := cpf.Format(fs.Arg(0))
		if err != nil {
			fmt.Fprintf(stderr, "docval cpf formatar: %s\n", err)
			return 1
		}
		fmt.Fprintln(stdout, formatado)
		return 0

	default:
		fmt.Fprintf(stderr, "docval cpf: subação desconhecida %q (use validar, formatar ou gerar)\n", action)
		return 2
	}
}
