// src/pages/QRCodePage.jsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { QRCodeSVG } from 'qrcode.react';
import api from '../services/api';
import jsPDF from 'jspdf';
import html2canvas from 'html2canvas';
import CustomLabelModal from '../components/CustomLabelModal'; // --- IMPORTADO ---

// O componente ReprintModal permanece o mesmo
const ReprintModal = ({ isOpen, onClose }) => {
    const [searchId, setSearchId] = useState('');
    const [foundCode, setFoundCode] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const handleSearch = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError('');
        setFoundCode(null);
        try {
            const response = await api.get(`/api/qrcodes/find/${searchId}`);
            if (response.data.data && response.data.data.length > 0) {
                setFoundCode(response.data.data[0]);
            }
        } catch (err) {
            setError(err.response?.data?.error || 'Falha ao buscar o código.');
        } finally {
            setLoading(false);
        }
    };

    const handlePrint = () => {
        window.print();
    };

    const handleClose = () => {
        setSearchId('');
        setFoundCode(null);
        setError('');
        onClose();
    };

    if (!isOpen) return null;

        const handleOverlayClick = (e) => {
        if (e.target === e.currentTarget) {
            handleClose();
        }
    };

    return (
        <div className="modal-overlay" onMouseDown={handleOverlayClick}>
            <div className="modal-content" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>Buscar e Reimprimir Etiqueta</h2>
                    <button className="modal-close-button" onClick={handleClose}>&times;</button>
                </div>
                <form onSubmit={handleSearch}>
                    <label>Código do STAGE IN</label>
                    <input
                        type="text"
                        value={searchId}
                        onChange={e => setSearchId(e.target.value.toUpperCase())}
                        placeholder="Ex: CG000123"
                        required
                    />
                    <button type="submit" className="modal-submit-button blue" disabled={loading}>
                        {loading ? 'Buscando...' : 'Buscar'}
                    </button>
                    {error && <p className="error-message">{error}</p>}
                </form>

                {foundCode && (
                    <div className="reprint-container">
                        <h3>Etiqueta Encontrada:</h3>
                        <div className="a4-page single-item-print">
                            <div className="qr-tag">
                                <div className="qr-tag-inner">
                                    <QRCodeSVG value={foundCode} size={128} includeMargin={true} />
                                    <p className="tag-id">{foundCode}</p>
                                    <p className="tag-stage">STAGE IN</p>
                                </div>
                            </div>
                        </div>
                        <button onClick={handlePrint} className="modal-submit-button entry">Imprimir Etiqueta</button>
                    </div>
                )}
            </div>
        </div>
    );
};

function QRCodePage() {
    const navigate = useNavigate();
    const [quantity, setQuantity] = useState(9);
    const [qrCodes, setQrCodes] = useState([]);
    const [isConfirmed, setIsConfirmed] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [isReprintModalOpen, setReprintModalOpen] = useState(false);
    // --- NOVO ESTADO ---
    const [isCustomLabelModalOpen, setCustomLabelModalOpen] = useState(false);

    const handleGenerate = async () => {
        if (quantity <= 0) {
            setError('Por favor, insira uma quantidade válida.');
            return;
        }
        setLoading(true);
        setError('');
        setQrCodes([]);
        setIsConfirmed(false);
        try {
            const response = await api.post('/api/qrcodes/generate-data', { quantity: parseInt(quantity, 10) });
            setQrCodes(response.data.data || []);
        } catch (err) {
            setError(err.response?.data?.error || 'Falha ao gerar os códigos.');
        } finally {
            setLoading(false);
        }
    };
    
    const handleConfirm = async () => {
        if (qrCodes.length === 0) return;
        setLoading(true);
        setError('');
        try {
            await api.post('/api/qrcodes/confirm', { trackingIds: qrCodes });
            setIsConfirmed(true);
        } catch (err) {
            setError(err.response?.data?.error || 'Falha ao confirmar os códigos.');
        } finally {
            setLoading(false);
        }
    };

    const handleExportPDF = async () => {
        if (qrCodes.length === 0) return;
        setLoading(true);
        const pdf = new jsPDF('p', 'mm', 'a4');
        const pageElements = document.querySelectorAll('.a4-page');
        for (let i = 0; i < pageElements.length; i++) {
            const page = pageElements[i];
            const canvas = await html2canvas(page, { scale: 2, useCORS: true, backgroundColor: '#ffffff' });
            const imgData = canvas.toDataURL('image/png');
            const pageWidth = pdf.internal.pageSize.getWidth();
            const pageHeight = pdf.internal.pageSize.getHeight();
            if (i > 0) {
                pdf.addPage();
            }
            pdf.addImage(imgData, 'PNG', 0, 0, pageWidth, pageHeight);
        }
        const firstId = qrCodes[0];
        const lastId = qrCodes[qrCodes.length - 1];
        const filename = `STAGE-IN ${firstId} A ${lastId}.pdf`;
        pdf.save(filename);
        setLoading(false);
    };

    const pages = [];
    for (let i = 0; i < qrCodes.length; i += 9) {
        pages.push(qrCodes.slice(i, i + 9));
    }

    return (
        <div className="qr-page-container">
            <div className="app-container logs-container non-printable">
                <header className="logs-header">
                    <h1>Gerador de Etiquetas</h1>
                    <button onClick={() => navigate('/')} className="back-button">Voltar</button>
                </header>
                
                <div className="controls-panel">
                    <div className="filters-panel">
                        <div className="input-with-label">
                            <label htmlFor="quantity">Quantidade de etiquetas</label>
                            <input
                                id="quantity"
                                type="number"
                                value={quantity}
                                onChange={(e) => setQuantity(e.target.value)}
                                min="1"
                            />
                        </div>

                        <button onClick={handleGenerate} disabled={loading} className="action-button entry">
                            {loading ? 'Gerando...' : 'Gerar Novos IDs'}
                        </button>
                        <button onClick={() => setReprintModalOpen(true)} className="modal-submit-button blue">
                            Buscar e Reimprimir
                        </button>
                        {/* --- NOVO BOTÃO --- */}
                        <button onClick={() => setCustomLabelModalOpen(true)} className="modal-submit-button blue">
                            Etiqueta Personalizada
                        </button>
                    </div>
                    {error && <p className="error-message">{error}</p>}
                    {qrCodes.length > 0 && (
                        <div className="confirmation-section">
                            {isConfirmed ? (
                                <p className="success-message">Códigos salvos com sucesso!</p>
                            ) : (
                                <p>Pré-visualização. Confirme para salvar os códigos no sistema.</p>
                            )}
                            <div className="confirmation-actions">
                                <button onClick={handleConfirm} disabled={loading || isConfirmed} className="action-button entry">
                                    {isConfirmed ? 'Salvo' : 'Confirmar e Salvar'}
                                </button>
                                <button 
                                    onClick={handleExportPDF} 
                                    disabled={!isConfirmed || loading} 
                                    className={`modal-submit-button blue ${!isConfirmed ? 'disabled' : ''}`}
                                >
                                    {loading ? 'Exportando...' : 'Exportar PDF'}
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>

            <div className="printable-area">
                {pages.map((pageCodes, pageIndex) => (
                    <div key={pageIndex} className="a4-page" id={`page-${pageIndex}`}>
                        {pageCodes.map(code => (
                            <div key={code} className="qr-tag">
                                <div className="qr-tag-inner">
                                    <QRCodeSVG value={code} size={128} includeMargin={true} />
                                    <p className="tag-id">{code}</p>
                                    <p className="tag-stage">STAGE IN</p>
                                </div>
                            </div>
                        ))}
                    </div>
                ))}
            </div>
            
            <ReprintModal isOpen={isReprintModalOpen} onClose={() => setReprintModalOpen(false)} />
            {/* --- RENDERIZAÇÃO DO NOVO MODAL --- */}
            <CustomLabelModal isOpen={isCustomLabelModalOpen} onClose={() => setCustomLabelModalOpen(false)} />
        </div>
    );
}

export default QRCodePage;