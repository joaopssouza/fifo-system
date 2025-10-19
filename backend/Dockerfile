# --- Estágio 1: Build ---
# Usa a imagem oficial do Go para compilar nosso código
FROM golang:1.25 AS builder

# Define o diretório de trabalho dentro do contêiner
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências
COPY go.mod go.sum ./

# Baixa as dependências. O Go as armazena em cache, então isso só roda quando o go.mod muda.
RUN go mod download

# Copia todo o código fonte do backend
COPY . .

# Compila a aplicação para um único binário estático.
# CGO_ENABLED=0 e GOOS=linux são essenciais para uma compilação limpa para o ambiente do contêiner.
RUN CGO_ENABLED=0 GOOS=linux go build -o /fifo-backend .

# --- Estágio 2: Produção ---
# Usa uma imagem base mínima. Alpine é pequena e segura.
FROM alpine:latest

# Instala o pacote 'tzdata' que contém as informações de fuso horário
# para que o Go possa encontrar "America/Sao_Paulo".
RUN apk add --no-cache tzdata

# Define o diretório de trabalho
WORKDIR /

# Copia APENAS o binário compilado do estágio de build.
# Isso resulta em uma imagem final extremamente pequena.
COPY --from=builder /fifo-backend /fifo-backend

# O comando que será executado quando o contêiner iniciar.
CMD ["/fifo-backend"]