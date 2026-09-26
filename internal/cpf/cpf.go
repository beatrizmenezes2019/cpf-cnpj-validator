// Package cpf implementa a validação, formatação e geração de números de CPF
// (Cadastro de Pessoas Físicas), o documento de identificação fiscal de
// pessoas físicas no Brasil.
package cpf

import (
	"errors"
	"math/rand"
	"regexp"
	"strings"
)

// ErrInvalidLength é retornado quando o CPF, após a remoção de caracteres de
// formatação, não possui exatamente 11 dígitos.
var ErrInvalidLength = errors.New("cpf: deve conter 11 dígitos")

// ErrAllDigitsEqual é retornado quando todos os dígitos são iguais (ex:
// "111.111.111-11"). Esses números são numericamente "válidos" pelo cálculo
// dos dígitos verificadores, mas são conhecidos por não serem CPFs reais.
var ErrAllDigitsEqual = errors.New("cpf: sequência de dígitos repetidos não é válida")

// ErrCheckDigits é retornado quando os dígitos verificadores não conferem.
var ErrCheckDigits = errors.New("cpf: dígitos verificadores inválidos")

var nonDigit = regexp.MustCompile(`\D`)

// onlyDigits remove qualquer caractere que não seja um dígito (pontos,
// hífen, espaços, etc.).
func onlyDigits(s string) string {
	return nonDigit.ReplaceAllString(s, "")
}

// Validate verifica se s é um CPF válido. A entrada pode estar formatada
// (com pontos e hífen) ou conter apenas os 11 dígitos.
func Validate(s string) error {
	digits := onlyDigits(s)

	if len(digits) != 11 {
		return ErrInvalidLength
	}

	if allDigitsEqual(digits) {
		return ErrAllDigitsEqual
	}

	d1 := checkDigit(digits[:9], 10)
	d2 := checkDigit(digits[:9]+string(d1), 11)

	if digits[9] != d1 || digits[10] != d2 {
		return ErrCheckDigits
	}

	return nil
}

// IsValid é um atalho para Validate que retorna apenas um booleano.
func IsValid(s string) bool {
	return Validate(s) == nil
}

// allDigitsEqual retorna true se todos os caracteres da string forem iguais.
func allDigitsEqual(digits string) bool {
	for i := 1; i < len(digits); i++ {
		if digits[i] != digits[0] {
			return false
		}
	}
	return true
}

// checkDigit calcula um dígito verificador do CPF a partir de base (os
// primeiros 9 ou 10 dígitos) usando peso decrescente a partir de
// firstWeight, conforme o algoritmo oficial da Receita Federal.
func checkDigit(base string, firstWeight int) byte {
	sum := 0
	weight := firstWeight
	for i := 0; i < len(base); i++ {
		n := int(base[i] - '0')
		sum += n * weight
		weight--
	}
	rest := sum % 11
	if rest < 2 {
		return '0'
	}
	return byte('0' + (11 - rest))
}

// Format recebe um CPF (com ou sem formatação) e devolve no formato
// "000.000.000-00". Retorna erro se a entrada não tiver 11 dígitos.
func Format(s string) (string, error) {
	digits := onlyDigits(s)
	if len(digits) != 11 {
		return "", ErrInvalidLength
	}
	var b strings.Builder
	b.WriteString(digits[0:3])
	b.WriteByte('.')
	b.WriteString(digits[3:6])
	b.WriteByte('.')
	b.WriteString(digits[6:9])
	b.WriteByte('-')
	b.WriteString(digits[9:11])
	return b.String(), nil
}

// Strip remove a formatação de um CPF, deixando apenas os 11 dígitos.
func Strip(s string) string {
	return onlyDigits(s)
}

// Generate gera um CPF válido aleatório, já formatado como
// "000.000.000-00". Útil para gerar massa de dados de teste.
func Generate() string {
	base := make([]byte, 9)
	for i := range base {
		base[i] = byte('0' + rand.Intn(10))
	}

	// Evita gerar acidentalmente uma sequência de dígitos repetidos.
	if allDigitsEqual(string(base) + string(base[0])) {
		base[0] = byte('0' + (int(base[0]-'0')+1)%10)
	}

	d1 := checkDigit(string(base), 10)
	d2 := checkDigit(string(base)+string(d1), 11)

	digits := string(base) + string(d1) + string(d2)
	formatted, _ := Format(digits)
	return formatted
}
