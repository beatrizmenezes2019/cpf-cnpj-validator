# ---- estágio de build ----
# Compila o binário em um ambiente completo com o toolchain do Go.
#
# O Trivy detectou 19 CVEs HIGH na stdlib embutida no binário quando
# compilado com a última patch da série 1.24 (1.24.13) — essa série mais
# antiga não recebe mais os backports de segurança. As correções só
# existem a partir de 1.25.13 ou 1.26.6. Usamos a tag flutuante "1.26"
# (sempre resolve para a última patch publicada) como piso; o `go.mod`
# declara "go 1.26.6" como versão mínima exigida — se a imagem trouxer
# algo um pouco mais antigo, o próprio Go baixa o patch exato sozinho
# (GOTOOLCHAIN=auto, o padrão desde o Go 1.21).
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Copia primeiro o go.mod para aproveitar o cache de camadas do Docker:
# só reexecuta o download de dependências se go.mod mudar.
COPY go.mod ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/docval ./cmd/docval

# ---- estágio final (runtime) ----
# Imagem distroless: sem shell, sem gerenciador de pacotes, sem libc extra.
# Reduz drasticamente a superfície de ataque e o número de CVEs relatados
# pelo scanner de vulnerabilidades (Trivy) no pipeline de CI/CD.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/docval /usr/local/bin/docval

USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/docval"]
CMD ["help"]
