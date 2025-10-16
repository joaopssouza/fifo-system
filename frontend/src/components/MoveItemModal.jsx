// src/components/MoveItemModal.jsx
import React, { useState, useEffect } from 'react';
import api from '../services/api';

function MoveItemModal({ isOpen, onClose, onSuccess, item }) {
    const [newRua, setNewRua] = useState('');
    const [error, setError] = useState('');

    useEffect(() => {
        if (item) {
            setNewRua(item.Rua);
        }
    }, [item]);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        if (!item) return;

        try {
            await api.put(`/api/package/move/${item.ID}`, { rua: newRua });
            onSuccess();
            onClose();
        } catch (err) {
            setError(err.response?.data?.error || 'Falha ao realocar o item.');
        }
    };

    if (!isOpen || !item) return null;

    // --- NOVA LÓGICA ---
    const handleOverlayClick = (e) => {
        if (e.target === e.currentTarget) {
            onClose();
        }
    };

    return (
        <div className="modal-overlay" onMouseDown={handleOverlayClick}>
            <div className="modal-content" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>Mover Item: {item.TrackingID}</h2>
                    <button className="modal-close-button" onClick={onClose}>&times;</button>
                </div>
                <form onSubmit={handleSubmit}>
                    <label htmlFor="newRua">Mudar rua do Item</label>
                    <input
                        id="newRua"
                        type="text"
                        value={newRua}
                        onChange={e => setNewRua(e.target.value)}
                        placeholder="Ex: RTS-003"
                        required
                    />
                    <button type="submit" className="modal-submit-button blue">Realocar</button>
                    {error && <p className="error-message">{error}</p>}
                </form>
            </div>
        </div>
    );
}

export default MoveItemModal;