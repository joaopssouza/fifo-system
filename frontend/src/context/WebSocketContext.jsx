// src/context/WebSocketContext.jsx
import React, { createContext, useState, useEffect, useContext, useCallback, useMemo } from 'react';
import { useAuth } from './AuthContext';

const WebSocketContext = createContext();

export const WebSocketProvider = ({ children }) => {
    const { token, user, isGuest } = useAuth();
    const [onlineUsers, setOnlineUsers] = useState([]);
    const [wsQueue, setWsQueue] = useState([]);
    const [wsBacklog, setWsBacklog] = useState(0);
    const [wsBufferCounts, setWsBufferCounts] = useState({ RTS: 0, EHA: 0, SAL: 0 });
    const [isConnected, setIsConnected] = useState(false);

    useEffect(() => {
        let ws = null;

        // Se houver token (utilizador autenticado)
        if (token) {
            const baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
            const wsURL = baseURL.replace(/^http/, 'ws');
            const finalWsUrl = `${wsURL}/api/ws?token=${token}`;

            console.log("Tentando conectar WebSocket (Autenticado):", finalWsUrl);
            ws = new WebSocket(finalWsUrl);

            ws.onopen = () => {
                console.log("Conexão WebSocket Estabelecida (Autenticado)");
                setIsConnected(true);
            };

            ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    console.log("Mensagem WS recebida:", message.type);

                    if (message.type === 'online_users') {
                        if (user && (user.role === 'admin' || user.role === 'leader')) {
                            setOnlineUsers(message.data || []);
                        }
                    } else if (message.type === 'queue_update') {
                        setWsQueue(message.queue || []);
                        setWsBacklog(message.backlog || 0);
                        setWsBufferCounts(message.bufferCounts || { RTS: 0, EHA: 0, SAL: 0 });
                    }
                } catch (e) {
                    console.error("Erro ao processar mensagem do WebSocket:", e);
                }
            };

            ws.onclose = (event) => {
                console.log("Conexão WebSocket Fechada (Autenticado)", event.reason);
                setIsConnected(false);
                // Limpa os dados ao desconectar
                setOnlineUsers([]);
                setWsQueue([]);
                setWsBacklog(0);
                setWsBufferCounts({ RTS: 0, EHA: 0, SAL: 0 });
            };
            ws.onerror = (error) => {
                console.error("Erro no WebSocket (Autenticado):", error);
                setIsConnected(false);
            };

        // --- LÓGICA PARA CONVIDADO ---
        } else if (isGuest) {
            console.log("Modo Convidado: WebSocket não será conectado. Simulando estado 'conectado'.");
            // Limpa estados que dependem do WS
            setOnlineUsers([]);
            setWsQueue([]);
            setWsBacklog(0);
            setWsBufferCounts({ RTS: 0, EHA: 0, SAL: 0 });
            // Define como "conectado" para que o Dashboard não mostre a mensagem
            // e use os dados da API inicial.
            setIsConnected(false); // <<-- SIMULA CONEXÃO PARA CONVIDADO

        // Se não for autenticado nem convidado (ex: na página de login)
        } else {
            setIsConnected(false);
            setOnlineUsers([]);
            setWsQueue([]);
            setWsBacklog(0);
            setWsBufferCounts({ RTS: 0, EHA: 0, SAL: 0 });
        }

        // Função de limpeza
        return () => {
            if (ws) {
                console.log("Fechando conexão WebSocket...");
                ws.close();
                setIsConnected(false); // Garante reset ao sair
            }
        };
    }, [token, isGuest, user]); // Depende de token, isGuest e user

    const value = useMemo(() => ({
        onlineUsers,
        wsQueue,
        wsBacklog,
        wsBufferCounts,
        isConnected
    }), [onlineUsers, wsQueue, wsBacklog, wsBufferCounts, isConnected]);

    return (
        <WebSocketContext.Provider value={value}>
            {children}
        </WebSocketContext.Provider>
    );
};

export const useWebSocket = () => {
    return useContext(WebSocketContext);
};