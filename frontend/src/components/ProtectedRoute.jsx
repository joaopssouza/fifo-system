// src/components/ProtectedRoute.jsx
import React from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const ProtectedRoute = ({ requiredPermission }) => {
    const { isAuthenticated, isLoading, hasPermission } = useAuth();

    if (isLoading) {
        return <div>Carregando...</div>;
    }

    if (!isAuthenticated) {
        return <Navigate to="/login" replace />;
    }
    
    if (requiredPermission && !hasPermission(requiredPermission)) {
        return <Navigate to="/" replace />; 
    }

    return <Outlet />;
};

export default ProtectedRoute;