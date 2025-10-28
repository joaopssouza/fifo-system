// src/context/WebSocketContext.jsx
import React, { createContext, useState, useEffect, useContext, useCallback, useMemo } from 'react';
import { useAuth } from './AuthContext';

const WebSocketContext = createContext();

export const WebSocketProvider = ({ children }) => {
    const { token, user, isGuest } = useAuth();
    const [onlineUsers, setOnlineUsers] = useState([]);
    const [wsQueue, setWsQueue] = useState([]);
    const [wsBacklog, setWsBacklog] = useState(0);
    const [wsBacklogValue, setWsBacklogValue] = useState(0);
    const [wsBufferCounts, setWsBufferCounts] = useState({ RTS: 0, EHA: 0, SAL: 0 });
    const [wsBufferValues, setWsBufferValues] = useState({ RTS: 0, EHA: 0, SAL: 0 });
    // --- NOVOS ESTADOS PARA TEMPO MÉDIO ---
    const [wsBufferAvgTimes, setWsBufferAvgTimes] = useState({ RTS: 0.0, EHA: 0.0 });
    const [isConnected, setIsConnected] = useState(false);

    useEffect(() => {
        let ws = null;

        if (token) {
            // ... (conexão ws) ...
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
                        setWsBacklog(message.backlogCount || 0);
                        setWsBacklogValue(message.backlogValue || 0);
                        setWsBufferCounts(message.bufferCounts || { RTS: 0, EHA: 0, SAL: 0 });
                        setWsBufferValues(message.bufferValues || { RTS: 0, EHA: 0, SAL: 0 });
                        // --- ARMAZENA TEMPO MÉDIO ---
                        setWsBufferAvgTimes(message.bufferAvgTimes || { RTS: 0.0, EHA: 0.0 });
                    }
                } catch (e) {
                    console.error("Erro ao processar mensagem do WebSocket:", e);
                }
            };

            ws.onclose = (event) => {
                console.log("Conexão WebSocket Fechada (Autenticado)", event.reason);
                setIsConnected(false);
                setOnlineUsers([]);
                setWsQueue([]);
                setWsBacklog(0);
                setWsBacklogValue(0);
                setWsBufferCounts({ RTS: 0, EHA: 0, SAL: 0 });
                setWsBufferValues({ RTS: 0, EHA: 0, SAL: 0 });
                // --- RESET TEMPO MÉDIO ---
                setWsBufferAvgTimes({ RTS: 0.0, EHA: 0.0 });
            };
            ws.onerror = (error) => {
                console.error("Erro no WebSocket (Autenticado):", error);
                setIsConnected(false);
            };

        } else if (isGuest) {
            console.log("Modo Convidado: WebSocket não será conectado.");
            setOnlineUsers([]);
            setWsQueue([]);
            setWsBacklog(0);
            setWsBacklogValue(0);
            setWsBufferCounts({ RTS: 0, EHA: 0, SAL: 0 });
            setWsBufferValues({ RTS: 0, EHA: 0, SAL: 0 });
             // --- RESET TEMPO MÉDIO ---
            setWsBufferAvgTimes({ RTS: 0.0, EHA: 0.0 });
            setIsConnected(false);

        } else {
            setIsConnected(false);
            setOnlineUsers([]);
            setWsQueue([]);
            setWsBacklog(0);
            setWsBacklogValue(0);
            setWsBufferCounts({ RTS: 0, EHA: 0, SAL: 0 });
            setWsBufferValues({ RTS: 0, EHA: 0, SAL: 0 });
             // --- RESET TEMPO MÉDIO ---
            setWsBufferAvgTimes({ RTS: 0.0, EHA: 0.0 });
        }

        return () => {
            if (ws) {
                console.log("Fechando conexão WebSocket...");
                ws.close();
                setIsConnected(false);
            }
        };
    }, [token, isGuest, user]);

    const value = useMemo(() => ({
        onlineUsers,
        wsQueue,
        wsBacklog,
        wsBacklogValue,
        wsBufferCounts,
        wsBufferValues,
        wsBufferAvgTimes, // <-- EXPOSTO
        isConnected
    }), [onlineUsers, wsQueue, wsBacklog, wsBacklogValue, wsBufferCounts, wsBufferValues, wsBufferAvgTimes, isConnected]); // <-- ATUALIZADO

    return (
        <WebSocketContext.Provider value={value}>
            {children}
        </WebSocketContext.Provider>
    );
};

export const useWebSocket = () => {
    return useContext(WebSocketContext);
};