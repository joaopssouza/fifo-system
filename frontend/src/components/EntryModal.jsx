// src/components/EntryModal.jsx
import React, { useState } from 'react';
import api from '../services/api';

function EntryModal({ isOpen, onClose, onSuccess }) {
    const [trackingId, setTrackingId] = useState('');
    const [buffer, setBuffer] = useState('');
    const [rua, setRua] = useState('');
    const [error, setError] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        try {
            await api.post('/api/entry', { trackingId, buffer, rua });
            onSuccess(); // Atualiza os dados no dashboard
            onClose();   // Fecha o modal
            // Limpa os campos
            setTrackingId('');
            setBuffer('');
            setRua('');
        } catch (err) {
            setError(err.response?.data?.error || 'Falha ao adicionar item.');
        }
    };

    if (!isOpen) return null;

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal-content" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>Nova Entrada FIFO</h2>
                    <button className="modal-close-button" onClick={onClose}>&times;</button>
                </div>
                <form onSubmit={handleSubmit}>
                    <label htmlFor="trackingId">ID do Item</label>
                    <input id="trackingId" type="text" value={trackingId} onChange={e => setTrackingId(e.target.value)} placeholder="Ex: CG02" required />

                    <label htmlFor="buffer">Buffer</label>
                    <input id="buffer" type="text" value={buffer} onChange={e => setBuffer(e.target.value)} placeholder="Ex: RTS" required />

                    <label htmlFor="rua">Rua</label>
                    <input id="rua" type="text" value={rua} onChange={e => setRua(e.target.value)} placeholder="Ex: RTS-002" required />

                    <button type="submit" className="modal-submit-button blue">Adicionar à Fila</button>
                    {error && <p className="error-message">{error}</p>}
                </form>
            </div>
        </div>
    );
}

export default EntryModal;