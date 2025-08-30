// src/App.jsx
import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';

import LoginPage from './pages/LoginPage';
import DashboardPage from './pages/DashboardPage';
import LogsPage from './pages/LogsPage';
import ProtectedRoute from './components/ProtectedRoute';

function App() {
    return (
        <AuthProvider>
            <Routes>
                <Route path="/login" element={<LoginPage />} />

                {/* Rotas Protegidas */}
                <Route element={<ProtectedRoute />}>
                    <Route path="/" element={<DashboardPage />} />
                </Route>

                {/* Rotas de Admin */}
                <Route element={<ProtectedRoute allowedRoles={['admin']} />}>
                    <Route path="/logs" element={<LogsPage />} />
                </Route>

            </Routes>
        </AuthProvider>
    );
}

export default App;