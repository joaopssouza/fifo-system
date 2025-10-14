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

        if (!buffer) {
            setError('Por favor, selecione um buffer.');
            return;
        }

        try {
            await api.post('/api/entry', { trackingId, buffer, rua });
            onSuccess();
            onClose();
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

                    <label>Buffer</label>
                    <div className="buffer-options">
                        <button
                            type="button"
                            className={`buffer-button ${buffer === 'RTS' ? 'selected' : ''}`}
                            onClick={() => setBuffer('RTS')}
                        >
                            RTS
                        </button>
                        <button
                            type="button"
                            className={`buffer-button ${buffer === 'EHA' ? 'selected' : ''}`}
                            onClick={() => setBuffer('EHA')}
                        >
                            EHA
                        </button>
                        <button
                            type="button"
                            className={`buffer-button ${buffer === 'SALVADOS' ? 'selected' : ''}`}
                            onClick={() => setBuffer('SALVADOS')}
                        >
                            SALVADOS
                        </button>
                    </div>

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