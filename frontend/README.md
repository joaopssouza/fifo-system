# FIFO System - Frontend ⚛️

Este diretório contém o código-fonte da aplicação Frontend do **FIFO System**, uma Single-Page Application (SPA) moderna e reativa, construída com **React** e **Vite**.

## ✨ Arquitetura e Estrutura

O frontend foi projetado para ser modular, escalável e de fácil manutenção, seguindo uma arquitetura baseada em componentes.

  * **`src/`**: Diretório principal que contém todo o código-fonte da aplicação.

      * **`components/`**: Contém componentes React reutilizáveis que formam os blocos de construção da UI, como modais, botões e cartões de métricas.
      * **`context/`**: Armazena os Contextos React, que gerem o estado global da aplicação.
          * `AuthContext.jsx`: Gere a autenticação, o estado do utilizador (logado ou convidado), o token JWT e as permissões.
          * `WebSocketContext.jsx`: Estabelece e gere a conexão WebSocket para receber dados em tempo real, como a lista de utilizadores online.
      * **`pages/`**: Contém os componentes que representam as páginas completas da aplicação, como o Dashboard, a página de Login e o painel de Administração.
      * **`services/`**: Centraliza a comunicação com a API backend.
          * `api.js`: Contém a instância do Axios pré-configurada com a `baseURL` da API, lida a partir das variáveis de ambiente.
      * **`App.jsx`**: Componente raiz que define o roteamento da aplicação utilizando `react-router-dom`.
      * **`main.jsx`**: Ponto de entrada da aplicação, onde o React é renderizado no DOM.

  * **`public/`**: Contém os ativos estáticos que são servidos diretamente, como o `index.html`, favicons e o manifesto da PWA.

  * **Ficheiros de Configuração:**

      * `vite.config.js`: Ficheiro de configuração do Vite, usado para definir plugins e configurar o servidor de desenvolvimento.
      * `package.json`: Lista as dependências do projeto e os scripts (`dev`, `build`, `lint`).
      * `.env`: Ficheiro para variáveis de ambiente locais, como a URL da API backend. **Este ficheiro não deve ser enviado para o repositório**.
      * `vercel.json`: Ficheiro de configuração específico para o deploy na Vercel. Contém regras de reescrita para garantir que o roteamento do React funcione corretamente em uma SPA.
      * `Dockerfile`: Define os passos para construir uma imagem Docker do frontend, utilizando Nginx para servir os ficheiros estáticos.

-----

## ☁️ Deploy na Vercel

A Vercel é a plataforma escolhida para o deploy do frontend devido à sua integração perfeita com repositórios Git e à sua CDN global de alta performance.

1.  **Conexão com o Repositório:** O projeto na Vercel é ligado diretamente ao repositório GitHub do FIFO System.
2.  **Build e Deploy Automático:** A cada `git push`, a Vercel aciona automaticamente o comando `npm run build`.
3.  **Distribuição Global:** Os ficheiros estáticos resultantes do build (localizados na pasta `dist`) são distribuídos pela CDN global da Vercel, garantindo que os utilizadores em qualquer parte do mundo tenham acesso rápido à aplicação.
4.  **Variáveis de Ambiente:** A variável `VITE_API_URL` é configurada diretamente no painel da Vercel, apontando para a URL do backend no Google Cloud Run.

-----

## 🚀 Como Executar Localmente

### ✔️ Pré-requisitos

  * **Node.js** e **npm** instalados.
  * A API backend deve estar em execução.

### ✔️ Passos

1.  **Navegue até a pasta do frontend:**

    ```bash
    cd fifo-system/frontend
    ```

2.  **Instale as dependências:**

    ```bash
    npm install
    ```

3.  **Configure o ambiente local:**

      * Crie um ficheiro `.env` na pasta `frontend`.
      * Adicione a seguinte linha, apontando para a sua API local:
        ```env
        VITE_API_URL="http://localhost:8080"
        ```

4.  **Inicie o servidor de desenvolvimento:**

    ```bash
    npm run dev
    ```

A aplicação estará disponível em [http://localhost:5173](https://www.google.com/search?q=http://localhost:5173) (ou na porta que for indicada no terminal).