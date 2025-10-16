// src/pages/DashboardPage.jsx
import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

import EntryModal from '../components/EntryModal';
import ExitModal from '../components/ExitModal';
import ChangePasswordModal from '../components/ChangePasswordModal';
import MoveItemModal from '../components/MoveItemModal';

const formatDuration = (seconds) => {
    if (isNaN(seconds) || seconds < 0) return '00:00:00';
    const h = Math.floor(seconds / 3600).toString().padStart(2, '0');
    const m = Math.floor((seconds % 3600) / 60).toString().padStart(2, '0');
    const s = Math.floor(seconds % 60).toString().padStart(2, '0');
    return `${h}:${m}:${s}`;
};

function DashboardPage() {
    // --- LINHA CORRIGIDA ---
    // A variável 'isGuest' foi adicionada à desestruturação do useAuth()
    const { user, logout, hasPermission, isGuest } = useAuth();
    // --- FIM DA CORREÇÃO ---

    const navigate = useNavigate();
    const [queue, setQueue] = useState([]);
    const [backlog, setBacklog] = useState(0);
    const [loading, setLoading] = useState(true);
    const [isEntryModalOpen, setEntryModalOpen] = useState(false);
    const [isExitModalOpen, setExitModalOpen] = useState(false);
    const [isChangePasswordModalOpen, setChangePasswordModalOpen] = useState(false);
    const [isMoveModalOpen, setMoveModalOpen] = useState(false);
    const [selectedItem, setSelectedItem] = useState(null);
    const [timeOffset, setTimeOffset] = useState(0);
    const [syncedTime, setSyncedTime] = useState(new Date().getTime());

    const fetchData = useCallback(async () => {
        try {
            const queueEndpoint = isGuest ? '/public/fifo-queue' : '/api/fifo-queue';
            const backlogEndpoint = isGuest ? '/public/backlog-count' : '/api/backlog-count';

            const [queueRes, backlogRes] = await Promise.all([
                api.get(queueEndpoint),
                api.get(backlogEndpoint)
            ]);
            setQueue(queueRes.data.data || []);
            setBacklog(backlogRes.data.count || 0);
        } catch (error) {
            console.error("Failed to fetch data", error);
        } finally {
            setLoading(false);
        }
    }, [isGuest]);

    useEffect(() => {
        fetchData();
        const syncTime = async () => {
            try {
                const response = await api.get('/public/time');
                const serverTime = new Date(response.data.serverTime).getTime();
                const localTime = new Date().getTime();
                setTimeOffset(serverTime - localTime);
            } catch (error) {
                console.error("Falha ao sincronizar o tempo com o servidor:", error);
                setTimeOffset(0);
            }
        };
        syncTime();
    }, [fetchData]);

    useEffect(() => {
        const interval = setInterval(() => {
            setSyncedTime(new Date().getTime() + timeOffset);
        }, 1000);
        return () => clearInterval(interval);
    }, [timeOffset]);

    const bufferCounts = useMemo(() => {
        return {
            RTS: queue.filter(item => item.Buffer === 'RTS').length,
            EHA: queue.filter(item => item.Buffer === 'EHA').length,
            SAL: queue.filter(item => item.Buffer === 'SAL').length,
        };
    }, [queue]);

    const oldestItemDuration = useMemo(() => {
        if (queue.length > 0) {
            const oldestItemTimestamp = new Date(queue[0].EntryTimestamp).getTime();
            return Math.floor((syncedTime - oldestItemTimestamp) / 1000);
        }
        return 0;
    }, [queue, syncedTime]);

    const openMoveModal = (item) => {
        setSelectedItem(item);
        setMoveModalOpen(true);
    };

    const closeMoveModal = () => {
        setSelectedItem(null);
        setMoveModalOpen(false);
    };

    if (loading) {
        return <p>Carregando...</p>;
    }

    return (
        <div className="app-container dashboard-container">
            <header className="dashboard-header">
                <div>
                    <h1>FIFO</h1>
                    <p>Sistema de Controle Logístico</p>
                </div>
                <div className="user-profile">
                    <span>{user?.username}</span>
                    {!isGuest && (
                        <button onClick={() => setChangePasswordModalOpen(true)} className="change-password-button">ALTERAR SENHA</button>
                    )}
                    <button onClick={logout} className="logout-button">SAIR</button>
                </div>
            </header>

            <main>
                <section className="metrics-grid">
                    <div className="metric-card">
                        <span className="metric-value">{backlog}</span>
                        <span className="metric-label">Itens na Fila</span>
                    </div>
                    <div className="metric-card">
                        <span className="metric-value">{formatDuration(oldestItemDuration)}</span>
                        <span className="metric-label">Maior Tempo</span>
                    </div>
                    <div className="metric-card buffer-card">
                        <div className="buffer-count">
                            <span>RTS:</span>
                            <span>{bufferCounts.RTS}</span>
                        </div>
                        <div className="buffer-count">
                            <span>EHA:</span>
                            <span>{bufferCounts.EHA}</span>
                        </div>
                        <div className="buffer-count">
                            <span>SALVADOS:</span>
                            <span>{bufferCounts.SAL}</span>
                        </div>
                    </div>
                </section>

                <section className="fifo-list">
                    {/* A classe 'with-actions' é adicionada apenas se NÃO for convidado */}
                    <header className={`fifo-list-header ${!isGuest ? 'with-actions' : ''}`}>
                        <span>ID</span>
                        <span>BUFFER</span>
                        <span>RUA</span>
                        <span>DURAÇÃO</span>
                        {/* A coluna de cabeçalho 'AÇÕES' é renderizada apenas se NÃO for convidado */}
                        {!isGuest && <span>AÇÕES</span>}
                    </header>
                    <div className="fifo-list-body">
                        {queue.length > 0 ? queue.map(item => {
                            const entryTimestamp = new Date(item.EntryTimestamp).getTime();
                            const durationSeconds = Math.floor((syncedTime - entryTimestamp) / 1000);
                            return (
                                // A classe 'with-actions' é adicionada apenas se NÃO for convidado
                                <div className={`fifo-list-item ${!isGuest ? 'with-actions' : ''}`} key={item.ID}>
                                    <span>{item.TrackingID}</span>
                                    <span>{item.Buffer}</span>
                                    <span>{item.Rua}</span>
                                    <span>{formatDuration(durationSeconds)}</span>
                                    {/* A célula de ações é renderizada apenas se NÃO for convidado */}
                                    {!isGuest && (
                                        <div className="action-buttons-cell">
                                            {hasPermission('MOVE_PACKAGE') && (
                                                <button onClick={() => openMoveModal(item)} className="move-btn">
                                                    Mover
                                                </button>
                                            )}
                                        </div>
                                    )}
                                </div>
                            );
                        }) : (
                            <p className="empty-queue-message">A fila está vazia.</p>
                        )}
                    </div>
                </section>

                {!isGuest && hasPermission('MANAGE_FIFO') && (
                    <section className="actions-grid">
                        <button className="action-button entry" onClick={() => setEntryModalOpen(true)}>ENTRADA</button>
                        <button className="action-button exit" onClick={() => setExitModalOpen(true)}>SAÍDA</button>
                    </section>
                )}

                <div className="admin-nav-buttons">
                    {!isGuest && hasPermission('GENERATE_QR_CODES') && (
                        <button onClick={() => navigate('/qrcode-generator')} className="admin-nav-button">
                            GERAR QR CODES
                        </button>
                    )}
                    {!isGuest && hasPermission('VIEW_LOGS') && (
                        <button onClick={() => navigate('/logs')} className="admin-nav-button">
                            VER LOGS DE ATIVIDADE
                        </button>
                    )}
                    {!isGuest && hasPermission('VIEW_USERS') && (
                        <button onClick={() => navigate('/management')} className="admin-nav-button">
                            PAINEL DE GESTÃO
                        </button>
                    )}
                </div>
            </main>

            <ChangePasswordModal isOpen={isChangePasswordModalOpen} onClose={() => setChangePasswordModalOpen(false)} />
            <EntryModal isOpen={isEntryModalOpen} onClose={() => setEntryModalOpen(false)} onSuccess={fetchData} />
            <ExitModal isOpen={isExitModalOpen} onClose={() => setExitModalOpen(false)} onSuccess={fetchData} availableIDs={queue.map(item => item.TrackingID)} />
            <MoveItemModal isOpen={isMoveModalOpen} onClose={closeMoveModal} onSuccess={fetchData} item={selectedItem} />
        </div>
    );
}

export default DashboardPage;