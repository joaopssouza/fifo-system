# FIFO System 📦

Um sistema de gestão de filas logísticas com interface **dark mode**, simples e funcional.

## 🔎 Descrição

O **FIFO System** é uma aplicação web para gerir em tempo real uma fila de controlo logístico. Permite:
- Entrada e saída de itens (com **reutilização de QR Codes**)
- Visualização do **backlog** e do **tempo de permanência**
- **Log de auditoria** completo com filtros avançados
- Gestão de utilizadores com perfis: **admin** e **fifo**

### 📄 Tecnologias Utilizadas
- **Go (Gin)** – API REST de alta performance  
- **React (Vite)** – Frontend SPA moderno e reativo  
- **PostgreSQL** – Base relacional robusta  
- **Docker & Docker Compose** – Dev/Prod totalmente contentorizado  

⚙️ **Aplicação dinâmica e completa**  
_Possui integração total com banco de dados e funcionalidades em tempo real._

---

## 🚀 Como Executar

### ✔️ Pré-requisitos
- **Docker** e **Docker Compose** instalados

### ✔️ Passos
1. **Clonar o repositório**
   ```bash
   git clone <url-do-seu-repositorio>
   cd fifo-system
   ```

2. **Criar o arquivo de ambiente**  
   No diretório `backend`, crie um arquivo chamado `.env` com o conteúdo:
   ```env
   # Credenciais do Banco de Dados
      POSTGRES_USER=SEU_USUARIO
      POSTGRES_PASSWORD=SUA_SENHA_AQUI
      POSTGRES_DB=SEU_BANCO_DE_DADOS
      HOST_DB=db

   # String de Conexão para a Aplicação Go

      DATABASE_URL="host=${HOST_DB} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} port=5432 sslmode=disable"

   # Segredo para os Tokens JWTs

      JWT_SECRET="SEU_SEGREDO_JWT_SUPER_SEGURO_AQUI"
   ```

3. **Subir a stack**
   ```bash
   docker-compose up --build
   ```

### 🌐 Link do Preview (local)
- **Frontend:** http://localhost:5173  
- **API:** http://localhost:8080  

---

## 👤 Primeiro Login

Ao iniciar a aplicação pela primeira vez com uma base de dados vazia, o sistema irá criar automaticamente um utilizador administrador padrão para si.

- **Utilizador:** `admin`
- **Senha:** `admin`

Aviso de Segurança: É crucial que altere esta senha padrão imediatamente após o seu primeiro login, utilizando a funcionalidade "Alterar Senha" no dashboard.