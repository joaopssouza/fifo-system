# --- Estágio 1: Build ---
# Usa a imagem oficial do Node.js para o build do React
FROM node:22-alpine AS builder

# Define o diretório de trabalho
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências
COPY package.json package-lock.json ./

# Instala as dependências
RUN npm install

# Copia todo o código fonte do frontend
COPY . .

# Executa o build de produção do React/Vite
RUN npm run build

# --- Estágio 2: Produção ---
# Usa uma imagem base de um servidor web leve e seguro (Nginx)
FROM nginx:stable-alpine

# Copia os arquivos estáticos gerados no build do React para a pasta que o Nginx serve
COPY --from=builder /app/dist /usr/share/nginx/html

# Expõe a porta 80, que é a porta padrão do Nginx
EXPOSE 80

# Quando o contêiner iniciar, o Nginx iniciará automaticamente para servir nossos arquivos.
CMD ["nginx", "-g", "daemon off;"]