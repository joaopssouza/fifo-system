// src/context/AuthContext.jsx
import React, { createContext, useState, useEffect, useContext } from 'react';
import { jwtDecode } from 'jwt-decode';
import api from '../services/api';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
    const [token, setToken] = useState(localStorage.getItem('token'));
    const [user, setUser] = useState(null);
    const [isLoading, setIsLoading] = useState(true); // <-- NOVO ESTADO DE CARREGAMENTO

    useEffect(() => {
        if (token) {
            try {
                const decoded = jwtDecode(token);
                if (decoded.exp * 1000 > Date.now()) {
                    setUser({ username: decoded.user, role: decoded.role });
                    api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
                } else {
                    logout();
                }
            } catch (error) {
                console.error("Failed to decode token", error);
                logout();
            }
        }
        // Ao final da verificação, independentemente do resultado, o carregamento termina.
        setIsLoading(false); // <-- FINALIZA O CARREGAMENTO
    }, [token]);

    const login = (newToken) => {
        localStorage.setItem('token', newToken);
        setToken(newToken);
    };

    const logout = () => {
        localStorage.removeItem('token');
        setToken(null);
        setUser(null);
        delete api.defaults.headers.common['Authorization'];
    };

    const isAuthenticated = !!token;

    // Expõe o novo estado 'isLoading' para o resto da aplicação
    return (
        <AuthContext.Provider value={{ isAuthenticated, token, user, isLoading, login, logout }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => {
    return useContext(AuthContext);
};
