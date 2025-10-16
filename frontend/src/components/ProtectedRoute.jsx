// src/components/ProtectedRoute.jsx
import React from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const ProtectedRoute = ({ requiredPermission }) => {
    const { isAuthenticated, isLoading, hasPermission, isGuest } = useAuth();

    if (isLoading) {
        return <div>Carregando...</div>;
    }

    // --- LÓGICA CORRIGIDA ---
    // Define claramente quem tem permissão para aceder à rota.
    // Um utilizador é permitido se estiver autenticado com um token OU se for um convidado.
    const isAllowed = isAuthenticated || isGuest;

    // Se o utilizador não for permitido de forma alguma, redireciona para o login.
    if (!isAllowed) {
        return <Navigate to="/login" replace />;
    }
    
    // Se for permitido, então verificamos as permissões específicas da rota.
    // Se a rota exige uma permissão que o utilizador não tem, redireciona para a página principal.
    if (requiredPermission && !hasPermission(requiredPermission)) {
        return <Navigate to="/" replace />; 
    }

    // Se passou por todas as verificações, renderiza o conteúdo da rota.
    return <Outlet />;
};

export default ProtectedRoute;