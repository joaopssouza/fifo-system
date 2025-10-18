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

# --- INÍCIO DAS MUDANÇAS ---

# Remove a configuração padrão do Nginx que usa a porta 80
RUN rm /etc/nginx/conf.d/default.conf

# Copia o nosso novo ficheiro de configuração para o template do Nginx
COPY nginx.conf /etc/nginx/templates/default.conf.template

# Expõe a porta 10000 (o padrão da Render)
EXPOSE 10000

# Este comando substitui a variável ${PORT} pela porta real e inicia o Nginx
CMD ["/bin/sh", "-c", "envsubst < /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/default.conf && nginx -g 'daemon off;'"]
# --- FIM DAS MUDANÇAS ---