// src/App.jsx
import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { WebSocketProvider } from './context/WebSocketContext'; // <-- PASSO 1: IMPORTAR

import LoginPage from './pages/LoginPage';
import DashboardPage from './pages/DashboardPage';
import LogsPage from './pages/LogsPage';
import AdminPage from './pages/AdminPage';
import ProtectedRoute from './components/ProtectedRoute';
import QRCodePage from './pages/QRCodePage';
import { SpeedInsights } from "@vercel/speed-insights/react"

function App() {
    return (
        // O AuthProvider continua a ser o provedor principal
        <AuthProvider>
            {/* PASSO 2: O WebSocketProvider envolve as rotas da aplicação */}
            {/* Agora, qualquer rota dentro dele pode aceder ao contexto do WebSocket */}
            <WebSocketProvider>
                <Routes>
                    <Route path="/login" element={<LoginPage />} />

                    {/* Rota principal, apenas requer autenticação */}
                    <Route element={<ProtectedRoute />}>
                        <Route path="/" element={<DashboardPage />} />
                    </Route>

                    {/* Rota de Logs, requer a permissão 'VIEW_LOGS' */}
                    <Route element={<ProtectedRoute requiredPermission="VIEW_LOGS" />}>
                        <Route path="/logs" element={<LogsPage />} />
                    </Route>

                    {/* Rota de Gestão, requer a permissão 'VIEW_USERS' */}
                    <Route element={<ProtectedRoute requiredPermission="VIEW_USERS" />}>
                        <Route path="/management" element={<AdminPage />} />
                    </Route>
                    
                    <Route element={<ProtectedRoute requiredPermission="GENERATE_QR_CODES" />}>
                        <Route path="/qrcode-generator" element={<QRCodePage />} />
                    </Route>
                </Routes>
                <SpeedInsights />
                <footer className="app-footer">
                    Criado por Joao Paulo S. S. (Ops62094)
                </footer>
            </WebSocketProvider>
        </AuthProvider>
    );
}

export default App;

