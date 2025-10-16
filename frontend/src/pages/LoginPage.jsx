// src/pages/LoginPage.jsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const styles = {
    loginContainer: {
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        textAlign: 'center',
        paddingTop: '5rem',
    },
    title: {
        fontSize: '3rem',
        fontWeight: 'bold',
        marginBottom: '0.5rem',
    },
    subtitle: {
        color: 'var(--color-secondary)',
        marginBottom: '3rem',
    },
    loginForm: {
        width: '100%',
    },
    loginButton: {
        backgroundColor: 'var(--color-accent-red)',
        color: 'white',
        marginTop: '1rem',
    },
    errorMessage: {
        color: 'var(--color-accent-red)',
        marginTop: '1rem',
    },
    guestLink: {
        background: 'none',
        border: 'none',
        color: 'var(--color-secondary)',
        textDecoration: 'underline',
        cursor: 'pointer',
        marginTop: '1.5rem',
        fontSize: '0.9rem'
    }
};

function LoginPage() {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();
    const auth = useAuth();

    const handleLogin = async (e) => {
        e.preventDefault();
        setError('');
        try {
            const response = await api.post('/login', { username, password });
            auth.login(response.data.token);
            navigate('/');
        } catch (err) {
            setError('Usuário ou senha inválidos.');
            console.error(err);
        }
    };

    const handleGuestLogin = () => {
        auth.loginAsGuest();
        navigate('/');
    };

    return (
        <div className="app-container" style={styles.loginContainer}>
            <h1 style={styles.title}>FIFO</h1>
            <p style={styles.subtitle}>Sistema de Controle Logístico</p>

            <form onSubmit={handleLogin} style={styles.loginForm}>
                <input
                    type="text"
                    placeholder="Usuário"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    required
                />
                <input
                    type="password"
                    placeholder="Senha"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                />
                <button type="submit" style={styles.loginButton}>
                    ENTRAR
                </button>
            </form>

            <button onClick={handleGuestLogin} style={styles.guestLink}>
                Entrar como convidado
            </button>
            
            {error && <p className="error-message">{error}</p>}
        </div>
    );
}

export default LoginPage;