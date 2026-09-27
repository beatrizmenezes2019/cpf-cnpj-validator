# docval

[![CI/CD Pipeline](https://github.com/beatrizmenezes2019/cpf-cnpj-validator/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/beatrizmenezes2019/cpf-cnpj-validator/actions/workflows/ci-cd.yml)

CLI em Go para validar, formatar e gerar números de CPF.

Feita para a disciplina de Implantação e Entrega de Software (Engenharia de
Software - UFG). Ver [`PIPELINE.md`](PIPELINE.md) para os detalhes da
pipeline de CI/CD com GitHub Actions.

## Uso

```bash
go build -o docval ./cmd/docval

./docval cpf validar 111.444.777-35
# 111.444.777-35: válido

./docval cpf validar 111.444.777-36
# 111.444.777-36: inválido (cpf: dígitos verificadores inválidos)
# (código de saída 1)

./docval cpf formatar 11144477735
# 111.444.777-35

./docval cpf gerar
# gera um CPF válido aleatório, ex: 643.923.477-37

./docval version
```

## Estrutura

```
cmd/docval/                 ponto de entrada da CLI (main.go)
internal/cpf/               lógica de validação, formatação e geração de CPF
Dockerfile                  build multi-stage da imagem (distroless)
.github/workflows/ci-cd.yml pipeline de CI/CD completa: verificação, build,
                             deploy em 3 ambientes e abertura automática de
                             PR/release (GitHub Actions)
PIPELINE.md                 documentação da pipeline e da automação
```

A lógica de negócio (`internal/cpf`) é composta só de funções puras, sem
efeitos colaterais, o que facilita os testes automatizados.

## Desenvolvimento

```bash
gofmt -l .          # formatação
go vet ./...        # análise estática
go test ./... -v -cover   # testes com cobertura
```

## Build com versão

A versão é injetada em tempo de build via `ldflags`, para uso posterior em
uma pipeline de CI/CD:

```bash
go build -ldflags "-X main.version=v1.0.0" -o docval ./cmd/docval
```

