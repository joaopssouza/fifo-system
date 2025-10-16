// src/context/AuthContext.jsx
import React, { createContext, useState, useEffect, useContext, useCallback, useMemo } from 'react';
import { jwtDecode } from 'jwt-decode';
import api from '../services/api';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
    const [token, setToken] = useState(() => localStorage.getItem('token'));
    const [user, setUser] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    // --- NOVO ESTADO ---
    const [isGuest, setIsGuest] = useState(() => sessionStorage.getItem('isGuest') === 'true');

    const logout = useCallback(() => {
        localStorage.removeItem('token');
        sessionStorage.removeItem('isGuest'); // --- ADICIONADO ---
        setToken(null);
        setUser(null);
        setIsGuest(false); // --- ADICIONADO ---
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
        // --- LÓGICA DE CONVIDADO ADICIONADA ---
        } else if (isGuest) {
            setUser({ username: 'Convidado', permissions: [] }); // Define um usuário "falso" para convidado
        } else {
            setUser(null);
        }
        setIsLoading(false);
    }, [token, logout, isGuest]);

    const login = useCallback((newToken) => {
        localStorage.setItem('token', newToken);
        sessionStorage.removeItem('isGuest');
        setIsGuest(false);
        setToken(newToken);
    }, []);

    // --- NOVA FUNÇÃO ---
    const loginAsGuest = useCallback(() => {
        localStorage.removeItem('token');
        sessionStorage.setItem('isGuest', 'true');
        setToken(null);
        setIsGuest(true);
    }, []);


    const hasPermission = useCallback((permissionName) => {
        // Convidados nunca têm permissões
        if (isGuest) return false;
        return user?.permissions.includes(permissionName) ?? false;
    }, [user, isGuest]);

    const isAuthenticated = !!token;

    const value = useMemo(() => ({
        isAuthenticated,
        isGuest, // --- EXPOSTO ---
        token,
        user,
        isLoading,
        login,
        loginAsGuest, // --- EXPOSTO ---
        logout,
        hasPermission
    }), [isAuthenticated, isGuest, token, user, isLoading, login, logout, hasPermission, loginAsGuest]);

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => {
    return useContext(AuthContext);
};