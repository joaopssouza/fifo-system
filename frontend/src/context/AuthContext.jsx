// src/context/AuthContext.jsx
import React, { createContext, useState, useEffect, useContext, useCallback, useMemo } from 'react';
import { jwtDecode } from 'jwt-decode';
import api from '../services/api';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
    const [token, setToken] = useState(() => localStorage.getItem('token'));
    const [user, setUser] = useState(null);
    const [isLoading, setIsLoading] = useState(true);

    const logout = useCallback(() => {
        localStorage.removeItem('token');
        setToken(null);
        setUser(null);
        delete api.defaults.headers.common['Authorization'];
    }, []);

    useEffect(() => {
        if (token) {
            try {
                const decoded = jwtDecode(token);
                if (decoded.exp * 1000 > Date.now()) {
                    setUser({
                        username: decoded.user,
                        FullName: decoded.fullName,
                        role: decoded.role,
                        permissions: decoded.permissions || []
                    });
                    api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
                } else {
                    logout();
                }
            } catch (error) {
                console.error("Falha ao descodificar o token", error);
                logout();
            }
        } else {
            setUser(null);
        }
        setIsLoading(false);
    }, [token, logout]);

    const login = useCallback((newToken) => {
        localStorage.setItem('token', newToken);
        setToken(newToken);
    }, []);

    const hasPermission = useCallback((permissionName) => {
        return user?.permissions.includes(permissionName) ?? false;
    }, [user]);

    const isAuthenticated = !!token;

    // Otimização crítica com useMemo para estabilizar o valor do contexto
    const value = useMemo(() => ({
        isAuthenticated,
        token,
        user,
        isLoading,
        login,
        logout,
        hasPermission
    }), [isAuthenticated, token, user, isLoading, login, logout, hasPermission]);

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => {
    return useContext(AuthContext);
};

