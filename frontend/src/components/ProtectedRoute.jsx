// src/components/ProtectedRoute.jsx
import React from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const ProtectedRoute = ({ allowedRoles }) => {
    const { isAuthenticated, user, isLoading } = useAuth(); // <-- OBTÉM O ESTADO DE CARREGAMENTO

    // 1. Se estiver a carregar, não renderiza nada ainda.
    if (isLoading) {
        return null; // Ou um componente de spinner/loading global
    }

    // 2. Após o carregamento, verifica se está autenticado.
    if (!isAuthenticated) {
        return <Navigate to="/login" replace />;
    }
    
    // 3. Verifica as permissões (roles).
    if (allowedRoles && !allowedRoles.includes(user?.role)) {
        return <Navigate to="/" replace />; 
    }

    // 4. Se tudo estiver correto, renderiza a página.
    return <Outlet />;
};

export default ProtectedRoute;
