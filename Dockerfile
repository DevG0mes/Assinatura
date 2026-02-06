# Etapa 1: Construção (Builder)
# Usa uma imagem leve do Go para compilar
FROM golang:1.23-alpine AS builder

# Define a pasta de trabalho dentro do container
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências
COPY go.mod go.sum ./

# Baixa as dependências (Gin, etc.)
RUN go mod download

# Copia todo o restante do código (main.go, assets, templates)
COPY . .

# Compila o executável chamado "main"
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Etapa 2: Execução (Runner)
# Usa uma imagem Linux mínima para rodar (ocupa menos espaço)
FROM alpine:latest

WORKDIR /app

# Copia o executável gerado na etapa anterior
COPY --from=builder /app/main .

# IMPORTANTE: Copia as pastas de templates e assets para o container final
# Sem isso, o site não acha as imagens nem o HTML
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/assets ./assets

# Expõe a porta 8080 (que é a que usamos no main.go)
EXPOSE 8080

# Comando para iniciar o servidor
CMD ["./main"]