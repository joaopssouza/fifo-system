// src/components/ChangePasswordModal.jsx
import React, { useState } from 'react';
import api from '../services/api';

function ChangePasswordModal({ isOpen, onClose }) {
    const [oldPassword, setOldPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        setSuccess('');

        if (newPassword.length < 6) {
            setError("A nova senha deve ter pelo menos 6 caracteres.");
            return;
        }

        try {
            const response = await api.put('/api/user/change-password', { oldPassword, newPassword });
            setSuccess(response.data.message);
            setOldPassword('');
            setNewPassword('');
            setTimeout(() => {
                onClose();
                setSuccess('');
            }, 2000);
        } catch (err) {
            setError(err.response?.data?.error || 'Não foi possível alterar a senha.');
        }
    };

    if (!isOpen) return null;

    // --- NOVA LÓGICA ---
    const handleOverlayClick = (e) => {
        // Fecha o modal apenas se o clique for diretamente no overlay
        if (e.target === e.currentTarget) {
            onClose();
        }
    };

    return (
        <div className="modal-overlay" onMouseDown={handleOverlayClick}>
            <div className="modal-content" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>Alterar Senha</h2>
                    <button className="modal-close-button" onClick={onClose}>&times;</button>
                </div>

                <form onSubmit={handleSubmit}>
                    <label htmlFor="oldPassword">Senha Atual</label>
                    <input id="oldPassword" type="password" value={oldPassword} onChange={e => setOldPassword(e.target.value)} required />
                    
                    <label htmlFor="newPassword">Nova Senha</label>
                    <input id="newPassword" type="password" value={newPassword} onChange={e => setNewPassword(e.target.value)} required />

                    <button type="submit" className="modal-submit-button blue">Confirmar Alteração</button>
                    
                    {error && <p className="error-message">{error}</p>}
                    {success && <p className="success-message">{success}</p>}
                </form>
            </div>
        </div>
    );
}

export default ChangePasswordModal;