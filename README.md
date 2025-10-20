# FIFO System 📦

Um sistema de gestão de filas logísticas de alta performance, com interface *dark mode*, simples e funcional, projetado para operar em uma arquitetura de nuvem moderna e escalável.

## 🎯 Objetivo do Projeto

O **FIFO System** foi desenvolvido para resolver a necessidade de um controle logístico eficiente e em tempo real. O seu principal objetivo é oferecer uma plataforma centralizada onde operadores e gestores possam gerir o fluxo de entrada e saída de itens (gaiolas) de um *buffer* seguindo a metodologia FIFO (*First-In, First-Out*).

A aplicação garante que os itens que entram primeiro sejam os primeiros a sair, otimizando o fluxo, reduzindo o tempo de permanência e fornecendo visibilidade completa sobre o estado da fila.

-----

## ✨ Funcionalidades Principais

  * **Gestão de Fila em Tempo Real:** Acompanhe a entrada, saída e movimentação de itens com atualizações instantâneas.
  * **Dashboard Intuitivo:** Visualize o *backlog* total, o tempo de permanência do item mais antigo e a contagem de itens por *buffer* (RTS, EHA, SALVADOS).
  * **Geração e Gestão de QR Codes:** Crie etiquetas de QR Code únicas para rastreamento de itens, com funcionalidades para reimpressão e geração de etiquetas personalizadas.
  * **Sistema de Autenticação e Permissões (RBAC):**
      * **Autenticação JWT:** Segurança baseada em tokens.
      * **Papéis (Roles):** Perfis de utilizador (`admin`, `leader`, `fifo`) com permissões distintas.
      * **Controlo de Acesso:** Ações críticas são restritas com base no papel do utilizador.
  * **Auditoria Completa:** Registo detalhado de todas as ações importantes (entrada, saída, movimentação) com filtros avançados por data, utilizador e tipo de ação.
  * **Gestão de Utilizadores:** Administradores e líderes podem criar, editar e redefinir senhas de outros utilizadores, respeitando uma hierarquia de permissões.
  * **Visualização de Utilizadores Online:** Líderes e administradores podem ver em tempo real quem está a utilizar o sistema, através de uma conexão WebSocket.
  * **Modo Convidado:** Permite a visualização do estado da fila sem necessidade de login, garantindo a transparência da operação para todos os interessados.

-----

## 🛠️ Tecnologias Utilizadas

| Componente | Tecnologia | Descrição |
| :--- | :--- | :--- |
| **Backend** | **Go (Gin)** | API REST de alta performance, otimizada para compilação em um binário estático e leve. |
| **Frontend** | **React (Vite)** | Uma Single-Page Application (SPA) moderna, reativa e de carregamento rápido. |
| **Base de Dados** | **PostgreSQL** | Sistema de gestão de base de dados relacional, robusto e confiável. |
| **WebSockets** | **Gorilla WebSocket** | Utilizado para comunicação bidirecional em tempo real (ex: lista de utilizadores online). |
| **Containerização** | **Docker** | O `Dockerfile` do backend prepara a aplicação para ser executada em qualquer ambiente de nuvem. |

-----

## ☁️ Arquitetura e Deploy na Nuvem

Este projeto foi desenhado para ser totalmente *cloud-native*, aproveitando serviços geridos para garantir escalabilidade, segurança e baixo custo operacional.

### 1\. Backend: Go em Google Cloud Run

  * **Como Funciona:** A API em Go é containerizada utilizando o `backend/Dockerfile`. Este Dockerfile cria uma imagem de contentor mínima e otimizada a partir do `alpine:latest`, contendo apenas o binário compilado da aplicação.
  * **Deploy:** Esta imagem é enviada para o Google Artifact Registry e, em seguida, implantada no **Google Cloud Run**. O Cloud Run executa o nosso contentor de forma *stateless* e escala automaticamente de zero a N instâncias conforme a demanda, o que o torna extremamente custo-eficiente.

### 2\. Frontend: React em Vercel

  * **Como Funciona:** O frontend desenvolvido com React e Vite é ligado diretamente a um repositório no GitHub.
  * **Deploy:** A **Vercel** faz o *build* automático do projeto a cada `git push`. Os ficheiros estáticos gerados (`HTML`, `CSS`, `JS`) são distribuídos globalmente através da sua CDN, garantindo tempos de carregamento ultra-rápidos para os utilizadores em qualquer lugar do mundo. O ficheiro `frontend/vercel.json` contém regras de reescrita para garantir que o roteamento do React (React Router) funcione corretamente.

### 3\. Base de Dados: PostgreSQL em Supabase

  * **Como Funciona:** Em vez de gerir uma máquina virtual para a base de dados, utilizamos a **Supabase**, que oferece instâncias PostgreSQL totalmente geridas.
  * **Conexão:** O backend em Go, a correr no Cloud Run, conecta-se de forma segura à base de dados na Supabase utilizando a `DATABASE_URL` fornecida nas variáveis de ambiente. Isto abstrai toda a complexidade de manutenção, backups e escalabilidade da base de dados.

-----

## 🔑 Configuração de Variáveis de Ambiente (`.env`)

O projeto **não utiliza um ficheiro `.env` na raiz**. A configuração é gerida por ficheiros `.env` individuais dentro das pastas `backend` e `frontend`, que são utilizados principalmente para desenvolvimento local. Em produção, estas variáveis são configuradas diretamente nos serviços de nuvem.

### `backend/.env`

Este ficheiro contém as variáveis de ambiente para a API em Go.

```env
# String de Conexão (Para a aplicação Go)
# Em produção, esta é a URL da sua instância PostgreSQL na Supabase.
DATABASE_URL="SUA_CONNECTION_STRING_SUPABASE"

# Segredo para assinatura do token (Algoritmo HS256 ou RS256).
# DEVE ser um valor longo, aleatório e mantido em segredo nas configurações do Cloud Run.
JWT_SECRET="SEU_SEGREDO_JWT_SUPER_SEGURO_AQUI"

# Tempo de expiração do token (Ex: 15m, 1h, 8h).
JWT_EXPIRATION_TIME="8h"

# Configurações da Aplicação
ENVIRONMENT="production"
GIN_MODE="release"

# Porta que o contentor irá expor (o Cloud Run gere a porta externa).
PORT="8080"

# URL do frontend para configurar o CORS (Cross-Origin Resource Sharing).
# Em produção, esta será a URL do seu site na Vercel.
FRONTEND_URL="https://SEU_PROJETO.vercel.app"
```

### `frontend/.env`

Este ficheiro contém as variáveis de ambiente para a aplicação React.

```env
# URL da sua API backend a correr no Google Cloud Run.
VITE_API_URL="https://SUA_API_BACKEND.a.run.app"
```

-----

## 🚀 Como Executar Localmente (com Docker)

Para desenvolvimento ou testes locais, pode utilizar o `docker-compose.yml` que simula a arquitetura da aplicação.

### ✔️ Pré-requisitos

  * **Docker** e **Docker Compose** instalados.

### ✔️ Passos

1.  **Clonar o repositório:**

    ```bash
    git clone <url-do-seu-repositorio>
    cd fifo-system
    ```

2.  **Configurar os ficheiros de ambiente:**

      * Crie e preencha o ficheiro `backend/.env` com as suas configurações locais (pode usar o `fifo-system/.env.example` como base).
      * Crie e preencha o ficheiro `frontend/.env` com a URL do seu backend local (ex: `VITE_API_URL="http://localhost:8080"`).
      * Crie um ficheiro `.env` na raiz do projeto apenas para o Docker Compose, com as credenciais da base de dados e do pgAdmin, usando o `fifo-system/.env.example` como referência.

3.  **Subir a stack:**

    ```bash
    docker-compose up --build
    ```

### 🌐 Acessos Locais:

  * **Frontend:** [http://localhost:5173](https://www.google.com/search?q=http://localhost:5173)
  * **API Backend:** [http://localhost:8080](https://www.google.com/search?q=http://localhost:8080)
  * **pgAdmin:** [http://localhost:5050](https://www.google.com/search?q=http://localhost:5050)

-----

## 👤 Primeiro Login

Ao iniciar a aplicação pela primeira vez com uma base de dados vazia, o sistema irá criar automaticamente um utilizador administrador padrão.

  * **Utilizador:** `admin`
  * **Senha:** `admin`

<<<<<<< HEAD
**Aviso de Segurança:** É crucial que altere esta senha padrão imediatamente após o seu primeiro login, utilizando a funcionalidade "Alterar Senha" no dashboard.
=======
**Aviso de Segurança:** É crucial que altere esta senha padrão imediatamente após o seu primeiro login, utilizando a funcionalidade "Alterar Senha" no dashboard.
>>>>>>> b0fd151a04e467687f1e92bf5649753fd03d3af2
