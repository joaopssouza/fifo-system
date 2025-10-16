// src/components/ExitModal.jsx
import React, { useState } from 'react';
import api from '../services/api';

function ExitModal({ isOpen, onClose, onSuccess, availableIDs }) {
    const [trackingId, setTrackingId] = useState('');
    const [error, setError] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        try {
            await api.post('/api/exit', { trackingId });
            onSuccess();
            onClose();
            setTrackingId('');
        } catch (err) {
            setError(err.response?.data?.error || 'Falha ao dar saída no item.');
        }
    };

    if (!isOpen) return null;
    
    const handleTagClick = (id) => {
        setTrackingId(id);
    };

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
                    <h2>Saída de Item</h2>
                    <button className="modal-close-button" onClick={onClose}>&times;</button>
                </div>
                <p className="modal-subtitle">Escaneie ou digite o ID manualmente</p>
                <form onSubmit={handleSubmit}>
                    <label htmlFor="exitTrackingId">ID do Produto</label>
                    <input id="exitTrackingId" type="text" value={trackingId} onChange={e => setTrackingId(e.target.value)} placeholder="Digite o ID ou escaneie" required />

                    <div className="available-ids">
                        <span>IDs sugeridos:</span>
                        {availableIDs.slice(0, 4).map(id => <span key={id} className="id-tag" onClick={() => handleTagClick(id)}>{id}</span>)}
                    </div>

                    <button type="submit" className="modal-submit-button red">Confirmar Saída</button>
                    {error && <p className="error-message">{error}</p>}
                </form>
            </div>
        </div>
    );
}

export default ExitModal;