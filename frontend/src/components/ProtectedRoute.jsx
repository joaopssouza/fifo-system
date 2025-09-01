// src/components/ProtectedRoute.jsx
import React from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

// --- COMPONENTE TOTALMENTE ATUALIZADO ---
const ProtectedRoute = ({ requiredPermission }) => {
    // Obtenha a função 'hasPermission' do contexto
    const { isAuthenticated, isLoading, hasPermission } = useAuth();

    // 1. Enquanto o estado de autenticação estiver a ser verificado, não renderiza nada.
    if (isLoading) {
        return null; // Evita o "piscar" da página
    }

    // 2. Se não estiver autenticado, redireciona para o login.
    if (!isAuthenticated) {
        return <Navigate to="/login" replace />;
    }
    
    // 3. Se uma permissão for necessária e o utilizador não a tiver, redireciona para o dashboard.
    // Esta é a lógica que resolve o bug.
    if (requiredPermission && !hasPermission(requiredPermission)) {
        return <Navigate to="/" replace />; 
    }

    // 4. Se tudo estiver correto, renderiza a página solicitada.
    return <Outlet />;
};

export default ProtectedRoute;
