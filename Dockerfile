# ---- estágio de build ----
# Compila o binário em um ambiente completo com o toolchain do Go.
FROM golang:1.24-alpine AS builder

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
