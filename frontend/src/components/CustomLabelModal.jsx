// src/components/CustomLabelModal.jsx
import React, { useState } from 'react';
import { QRCodeSVG } from 'qrcode.react';

function CustomLabelModal({ isOpen, onClose }) {
    const [prefix, setPrefix] = useState('RTS');
    const [identifier, setIdentifier] = useState('001');
    const [visibleLabel, setVisibleLabel] = useState('');

    const handleVisualize = (e) => {
        e.preventDefault();
        if (prefix && identifier) {
            setVisibleLabel(`${prefix}-${identifier}`);
        }
    };

    const handlePrint = () => {
        window.print();
    };

    const handleClose = () => {
        setPrefix('RTS');
        setIdentifier('001');
        setVisibleLabel('');
        onClose();
    };

    if (!isOpen) return null;

    // --- NOVA LÓGICA ---
    const handleOverlayClick = (e) => {
        if (e.target === e.currentTarget) {
            handleClose();
        }
    };

    return (
        <div className="modal-overlay" onMouseDown={handleOverlayClick}>
            <div className="modal-content" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>Gerar Etiqueta Personalizada</h2>
                    <button className="modal-close-button" onClick={handleClose}>&times;</button>
                </div>
                <form onSubmit={handleVisualize}>
                    <div className="custom-label-inputs">
                        <div className="input-wrapper">
                            <label htmlFor="prefix">Prefixo</label>
                            <input
                                id="prefix"
                                type="text"
                                value={prefix}
                                onChange={e => setPrefix(e.target.value.toUpperCase())}
                                placeholder="Ex: RTS"
                                required
                            />
                        </div>
                        <span className="input-separator">-</span>
                        <div className="input-wrapper">
                            <label htmlFor="identifier">Identificador</label>
                            <input
                                id="identifier"
                                type="text"
                                value={identifier}
                                onChange={e => setIdentifier(e.target.value)}
                                placeholder="Ex: 001"
                                required
                            />
                        </div>
                    </div>

                    <button type="submit" className="modal-submit-button blue">Visualizar</button>
                </form>

                {visibleLabel && (
                    <div className="reprint-container">
                        <h3>Pré-visualização da Etiqueta:</h3>
                        <div className="a4-page single-item-print">
                            <div className="qr-tag custom-label-tag">
                                <div className="qr-tag-inner">
                                    <QRCodeSVG value={visibleLabel} size={128} includeMargin={true} />
                                    <p className="tag-id">{visibleLabel}</p>
                                </div>
                            </div>
                        </div>
                        <button onClick={handlePrint} className="modal-submit-button entry">Imprimir Etiqueta</button>
                    </div>
                )}
            </div>
        </div>
    );
}

export default CustomLabelModal;