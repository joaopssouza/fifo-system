# FIFO System - Backend ⚙️

Este diretório contém o código-fonte da API REST de alta performance do **FIFO System**, desenvolvida em **Go** com o framework **Gin**. A API é responsável por toda a lógica de negócio, gestão de dados e comunicação em tempo real da aplicação.

## ✨ Arquitetura e Estrutura do Projeto

O backend segue uma arquitetura modular e organizada para garantir a separação de responsabilidades (*Separation of Concerns*), facilitando a manutenção e escalabilidade.

  * **`main.go`**: Ponto de entrada da aplicação. Responsável por inicializar a base de dados, configurar o servidor Gin, definir as rotas (públicas e privadas), aplicar middlewares e iniciar o servidor.
  * **`Dockerfile`**: Ficheiro de construção do contentor. Utiliza uma abordagem multi-stage para criar uma imagem Docker final extremamente leve e segura, baseada em `alpine:latest`, contendo apenas o binário compilado da aplicação.
  * **`/config`**: Gestão de configuração. Carrega variáveis de ambiente a partir de um ficheiro `.env` (para desenvolvimento) ou do ambiente de execução (para produção no Cloud Run).
  * **`/controllers`**: Contém a lógica de manipulação das requisições HTTP. Os controladores recebem os pedidos, validam os dados de entrada e orquestram as chamadas aos serviços para executar a lógica de negócio.
  * **`/initializers`**: Responsável por inicializar as conexões centrais da aplicação, como a ligação à base de dados PostgreSQL.
  * **`/middleware`**: Contém os middlewares do Gin.
      * `requireAuth.go`: Interceta as requisições a rotas protegidas, valida o token JWT e injeta os dados do utilizador no contexto da requisição.
      * `RequirePermission`: Garante que o utilizador autenticado possui a permissão específica necessária para aceder a um determinado *endpoint*.
  * **`/models`**: Define as estruturas de dados (entidades) que são mapeadas para as tabelas da base de dados utilizando o GORM. Inclui `User`, `Role`, `Permission`, `Package` e `AuditLog`.
  * **`/services`**: Centraliza a lógica de negócio reutilizável, como a criação de logs de auditoria e a gestão de tempo, garantindo consistência em toda a aplicação.
  * **`/websocket`**: Implementa a comunicação em tempo real utilizando WebSockets para funcionalidades como a lista de utilizadores online.

-----

## ☁️ Deploy em Google Cloud Run

O backend foi projetado para ser executado como um serviço *stateless* e escalável no Google Cloud Run.

1.  **Containerização:** O `Dockerfile` cria uma imagem de contentor otimizada. No primeiro estágio (`builder`), o código Go é compilado num único binário estático. No segundo estágio, apenas este binário e os dados de fuso horário (`tzdata`) são copiados para uma imagem base `alpine` mínima, resultando numa imagem final segura e com poucos megabytes.
2.  **Envio para o Artifact Registry:** A imagem Docker gerada é enviada para o Google Artifact Registry (o serviço de registo de contentores da Google Cloud).
3.  **Implantação no Cloud Run:** Um novo serviço é criado no Cloud Run, utilizando a imagem do Artifact Registry. O Cloud Run gere automaticamente o provisionamento, a escalabilidade (incluindo a redução a zero para economizar custos) e a exposição do serviço através de um URL HTTPS seguro.
4.  **Gestão de Segredos:** Todas as variáveis de ambiente sensíveis, como a `DATABASE_URL` e o `JWT_SECRET`, são configuradas de forma segura como "Segredos" no Secret Manager da Google Cloud e injetadas no ambiente de execução do Cloud Run.

-----

## 🔑 Configuração de Variáveis de Ambiente (`.env`)

Para desenvolvimento local, o backend utiliza um ficheiro `.env` na sua raiz. **Este ficheiro não deve ser enviado para o repositório.**

```env
# String de Conexão para a base de dados PostgreSQL.
# Para desenvolvimento local, aponta para o contentor Docker do Postgres.
# Em produção (Cloud Run), esta variável contém a connection string da Supabase.
DATABASE_URL="postgresql://USUARIO:SENHA@host:port/database"

# Segredo para assinatura do token JWT.
# Este valor DEVE ser longo, aleatório e mantido em segredo.
JWT_SECRET="SEU_SEGREDO_JWT_SUPER_SEGURO_AQUI"

# Tempo de expiração do token (formatos aceites: "15m", "1h", "8h").
JWT_EXPIRATION_TIME="8h"

# Modo de execução do Gin ("debug" para desenvolvimento, "release" para produção).
GIN_MODE="debug"

# Porta em que a API será executada dentro do contentor.
PORT="8080"

# URL do frontend para permitir requisições (CORS).
FRONTEND_URL="http://localhost:5173"
```

-----

## 🚀 Como Executar Localmente

### ✔️ Pré-requisitos

  * **Go (versão 1.25 ou superior)** instalado.
  * Uma instância **PostgreSQL** em execução (pode ser via Docker, conforme o `docker-compose.yml` na raiz do projeto).

### ✔️ Passos

1.  **Navegue até a pasta do backend:**

    ```bash
    cd fifo-system/backend
    ```

2.  **Instale as dependências:**

    ```bash
    go mod tidy
    ```

3.  **Configure o ambiente local:**

      * Crie um ficheiro `.env` na raiz da pasta `backend`.
      * Preencha-o com as suas configurações, apontando `DATABASE_URL` para a sua instância PostgreSQL e definindo um `JWT_SECRET`.

4.  **Execute a aplicação:**

    ```bash
    go run main.go
    ```

A API estará em execução em [http://localhost:8080](https://www.google.com/search?q=http://localhost:8080) (ou na porta definida na sua variável `PORT`).