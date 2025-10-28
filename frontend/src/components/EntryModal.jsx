// src/components/EntryModal.jsx
import React, { useState } from 'react';
import api from '../services/api';

function EntryModal({ isOpen, onClose, onSuccess }) {
    const [trackingId, setTrackingId] = useState('');
    const [buffer, setBuffer] = useState('');
    const [rua, setRua] = useState('');
    const [profile, setProfile] = useState('');
    const [error, setError] = useState('');

    // --- FUNÇÃO PARA LIMPAR ESTADOS ---
    const clearForm = () => {
        setTrackingId('');
        setBuffer('');
        setRua('');
        setProfile('');
        setError('');
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');

        if (!buffer) {
            setError('Por favor, selecione um buffer.');
            return;
        }
        // Validação de perfil só se não for SAL
        if (buffer !== 'SAL' && !profile) {
            setError('Por favor, selecione um perfil de pacote para RTS ou EHA.');
            return;
        }

        try {
            // Envia profile: null se for SAL, senão envia o selecionado
            const payload = {
                trackingId,
                buffer,
                rua,
                profile: buffer === 'SAL' ? null : profile
            };
            await api.post('/api/entry', payload);
            onSuccess();
            onClose(); // Fecha antes de limpar
            clearForm(); // Limpa depois de fechar
        } catch (err) {
            setError(err.response?.data?.error || 'Falha ao adicionar item.');
        }
    };

    // --- ATUALIZA LIMPEZA NO FECHAMENTO ---
    const handleClose = () => {
        clearForm(); // Limpa ao fechar pelo botão ou overlay
        onClose();
    }

    if (!isOpen) return null;

    const handleOverlayClick = (e) => {
        if (e.target === e.currentTarget) {
             handleClose(); // Usa a função handleClose
        }
    };

    return (
        <div className="modal-overlay" onMouseDown={handleOverlayClick}>
            <div className="modal-content" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>Nova Entrada FIFO</h2>
                    <button className="modal-close-button" onClick={handleClose}>&times;</button>
                </div>
                <form onSubmit={handleSubmit}>
                    <label htmlFor="trackingId">ID do Item</label>
                    <input id="trackingId" type="text" value={trackingId} onChange={e => setTrackingId(e.target.value)} placeholder="Ex: CG02" required />

                    <label>Buffer</label>
                    <div className="buffer-options">
                        {/* Atualiza o estado e limpa o perfil se mudar para SAL */}
                        <button type="button" className={`buffer-button ${buffer === 'RTS' ? 'selected' : ''}`} onClick={() => setBuffer('RTS')}> RTS </button>
                        <button type="button" className={`buffer-button ${buffer === 'EHA' ? 'selected' : ''}`} onClick={() => setBuffer('EHA')}> EHA </button>
                        <button type="button" className={`buffer-button ${buffer === 'SAL' ? 'selected' : ''}`} onClick={() => { setBuffer('SAL'); setProfile(''); }}> SALVADOS </button>
                    </div>

                    {/* --- SEÇÃO DE PERFIL CONDICIONAL --- */}
                    {buffer && buffer !== 'SAL' && (
                        <>
                            <label>Perfil do Pacote (Quantidade)</label>
                            <div className="buffer-options">
                                <button type="button" className={`buffer-button ${profile === 'P' ? 'selected' : ''}`} onClick={() => setProfile('P')}> P (250) </button>
                                <button type="button" className={`buffer-button ${profile === 'M' ? 'selected' : ''}`} onClick={() => setProfile('M')}> M (80) </button>
                                <button type="button" className={`buffer-button ${profile === 'G' ? 'selected' : ''}`} onClick={() => setProfile('G')}> G (10) </button>
                            </div>
                        </>
                    )}
                    {/* --- FIM DA SEÇÃO CONDICIONAL --- */}

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