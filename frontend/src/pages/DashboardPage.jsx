    // src/pages/DashboardPage.jsx
    import React, { useState, useEffect, useCallback } from 'react';
    import { useNavigate } from 'react-router-dom';
    import { useAuth } from '../context/AuthContext';
    import api from '../services/api';

    import EntryModal from '../components/EntryModal';
    import ExitModal from '../components/ExitModal';
    import ChangePasswordModal from '../components/ChangePasswordModal'; // <-- IMPORTA O NOVO MODAL

    // ... (função formatDuration existente)
    const formatDuration = (seconds) => {
        if (isNaN(seconds) || seconds < 0) return '00:00:00';
        const h = Math.floor(seconds / 3600).toString().padStart(2, '0');
        const m = Math.floor((seconds % 3600) / 60).toString().padStart(2, '0');
        const s = Math.floor(seconds % 60).toString().padStart(2, '0');
        return `${h}:${m}:${s}`;
    };


    function DashboardPage() {
        const { user, logout } = useAuth();
        const navigate = useNavigate();

        const [queue, setQueue] = useState([]);
        const [backlog, setBacklog] = useState(0);
        const [oldestItemDuration, setOldestItemDuration] = useState(0);
        const [loading, setLoading] = useState(true);
        
        const [isEntryModalOpen, setEntryModalOpen] = useState(false);
        const [isExitModalOpen, setExitModalOpen] = useState(false);
        const [isChangePasswordModalOpen, setChangePasswordModalOpen] = useState(false); // <-- NOVO ESTADO

        // ... (funções fetchData e useEffects existentes)
        const fetchData = useCallback(async () => {
            try {
                const [queueRes, backlogRes] = await Promise.all([
                    api.get('/api/fifo-queue'),
                    api.get('/api/backlog-count')
                ]);
                setQueue(queueRes.data.data || []);
                setBacklog(backlogRes.data.count || 0);
            } catch (error) {
                console.error("Failed to fetch data", error);
            } finally {
                setLoading(false);
            }
        }, []);

        useEffect(() => {
            fetchData();
        }, [fetchData]);

        useEffect(() => {
            const interval = setInterval(() => {
                if (queue.length > 0) {
                    const oldestItemTimestamp = new Date(queue[0].EntryTimestamp).getTime();
                    const now = new Date().getTime();
                    const durationInSeconds = Math.floor((now - oldestItemTimestamp) / 1000);
                    setOldestItemDuration(durationInSeconds);
                } else {
                    setOldestItemDuration(0);
                }
            }, 1000);

            return () => clearInterval(interval);
        }, [queue]);

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
                        {/* --- NOVO BOTÃO ADICIONADO --- */}
                        <button onClick={() => setChangePasswordModalOpen(true)} className="change-password-button">Alterar Senha</button>
                        <button onClick={logout} className="logout-button">Sair</button>
                    </div>
                </header>

                {/* ... (resto do JSX do dashboard) ... */}
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
                    </section>

                    <section className="fifo-list">
                        <header className="fifo-list-header">
                            <span>ID</span>
                            <span>BUFFER</span>
                            <span>RUA</span>
                            <span>DURAÇÃO</span>
                        </header>
                        <div className="fifo-list-body">
                            {queue.length > 0 ? queue.map(item => {
                                const durationSeconds = Math.floor((new Date().getTime() - new Date(item.EntryTimestamp).getTime()) / 1000);
                                return (
                                    <div className="fifo-list-item" key={item.ID}>
                                        <span>{item.TrackingID}</span>
                                        <span>{item.Buffer}</span>
                                        <span>{item.Rua}</span>
                                        <span>{formatDuration(durationSeconds)}</span>
                                    </div>
                                );
                            }) : (
                                <p className="empty-queue-message">A fila está vazia.</p>
                            )}
                        </div>
                    </section>

                    {user && (user.role === 'admin' || user.role === 'fifo') && (
                        <section className="actions-grid">
                            <button className="action-button entry" onClick={() => setEntryModalOpen(true)}>ENTRADA</button>
                            <button className="action-button exit" onClick={() => setExitModalOpen(true)}>SAÍDA</button>
                        </section>
                    )}
                    {user && user.role === 'admin' && (
                        <button onClick={() => navigate('/logs')} className="link-button">Ver Logs de Atividade</button>
                    )}
                </main>
                
                {/* --- RENDERIZA O NOVO MODAL --- */}
                <ChangePasswordModal
                    isOpen={isChangePasswordModalOpen}
                    onClose={() => setChangePasswordModalOpen(false)}
                />

                <EntryModal 
                    isOpen={isEntryModalOpen}
                    onClose={() => setEntryModalOpen(false)}
                    onSuccess={fetchData} 
                />
                <ExitModal 
                    isOpen={isExitModalOpen}
                    onClose={() => setExitModalOpen(false)}
                    onSuccess={fetchData}
                    availableIDs={queue.map(item => item.TrackingID)}
                />
            </div>
        );
    }

    export default DashboardPage;
    