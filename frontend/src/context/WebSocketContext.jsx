// src/context/WebSocketContext.jsx
import React, { createContext, useState, useEffect, useContext } from 'react';
import { useAuth } from './AuthContext';

const WebSocketContext = createContext();

export const WebSocketProvider = ({ children }) => {
    const { token, user } = useAuth();
    const [onlineUsers, setOnlineUsers] = useState([]);

    useEffect(() => {
        if (token && user && (user.role === 'admin' || user.role === 'leader')) {
            const baseURL = api.defaults.baseURL;

            // 3. Construa a URL do WebSocket de forma dinâmica
            const wsProtocol = baseURL.startsWith('https://') ? 'wss://' : 'ws://';
            const wsHost = baseURL.replace(/^https?:\/\//, '');
            const wsUrl = `${wsProtocol}${wsHost}/api/ws?token=${token}`; 
            const ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                console.log("Conexão WebSocket Estabelecida:", wsUrl);
            };

            ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    // A lógica de receção de dados não muda.
                    // Ela receberá tanto a resposta ao nosso pedido como as atualizações de outros clientes.
                    if (message.type === 'online_users') {
                        setOnlineUsers(message.data || []);
                    }
                } catch (e) {
                    console.error("Erro ao processar mensagem do WebSocket:", e);
                }
            };

            ws.onclose = () => console.log("Conexão WebSocket Fechada (WebSocketContext)");
            ws.onerror = (error) => console.error("Erro no WebSocket:", error);

            return () => {
                ws.close();
            };
        } else {
            setOnlineUsers([]);
        }
    }, [token, user]);

    return (
        <WebSocketContext.Provider value={{ onlineUsers }}>
            {children}
        </WebSocketContext.Provider>
    );
};

export const useWebSocket = () => {
    return useContext(WebSocketContext);
};

