// src/pages/DashboardPage.jsx
import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useWebSocket } from '../context/WebSocketContext';
import api from '../services/api';

import EntryModal from '../components/EntryModal';
import ExitModal from '../components/ExitModal';
import ChangePasswordModal from '../components/ChangePasswordModal';
import MoveItemModal from '../components/MoveItemModal';

const formatDuration = (seconds) => {
    // ... (função inalterada) ...
    if (isNaN(seconds) || seconds < 0) return '00:00:00';
    const h = Math.floor(seconds / 3600).toString().padStart(2, '0');
    const m = Math.floor((seconds % 3600) / 60).toString().padStart(2, '0');
    const s = Math.floor(seconds % 60).toString().padStart(2, '0');
    return `${h}:${m}:${s}`;
};

function DashboardPage() {
    const { user, logout, hasPermission, isGuest } = useAuth();
    // --- CONTEXTO ATUALIZADO ---
    const {
        wsQueue,
        wsBacklog, // Contagem Backlog
        wsBufferCounts, // Contagem Buffers
        isConnected,
        wsBacklogValue, // Soma Backlog
        wsBufferValues, // Soma Buffers
        wsBufferAvgTimes // Tempos Médios Buffers
    } = useWebSocket();
    const navigate = useNavigate();

    const [initialDataLoaded, setInitialDataLoaded] = useState(false);
    const [isEntryModalOpen, setEntryModalOpen] = useState(false);
    const [isExitModalOpen, setExitModalOpen] = useState(false);
    const [isChangePasswordModalOpen, setChangePasswordModalOpen] = useState(false);
    const [isMoveModalOpen, setMoveModalOpen] = useState(false);
    const [selectedItem, setSelectedItem] = useState(null);
    const [timeOffset, setTimeOffset] = useState(0);
    const [syncedTime, setSyncedTime] = useState(new Date().getTime());
    const [filterBuffer, setFilterBuffer] = useState('ALL');

    // --- ESTADOS DE FALLBACK ATUALIZADOS ---
    const [fallbackQueue, setFallbackQueue] = useState([]);
    const [fallbackBacklog, setFallbackBacklog] = useState(0);
    const [fallbackBacklogValue, setFallbackBacklogValue] = useState(0);
    const [fallbackCounts, setFallbackCounts] = useState({ RTS: 0, EHA: 0, SAL: 0 });
    const [fallbackValues, setFallbackValues] = useState({ RTS: 0, EHA: 0, SAL: 0 });
    const [fallbackAvgTimes, setFallbackAvgTimes] = useState({ RTS: 0.0, EHA: 0.0 }); // NOVO

    // --- API FALLBACK ATUALIZADA ---
    const fetchDataApi = useCallback(async () => {
        // ... (lógica de skip) ...
		if (initialDataLoaded) {
             console.log("fetchDataApi: Skipping, initial data already loaded.");
             return;
        }
        console.log("fetchDataApi: Buscando dados via API...");

        try {
            const queueEndpoint = isGuest ? '/public/fifo-queue' : '/api/fifo-queue';
            const backlogEndpoint = isGuest ? '/public/backlog-count' : '/api/backlog-count';
            const countsEndpoint = isGuest ? '/public/buffer-counts' : '/api/buffer-counts';

            const [queueRes, backlogRes, countsRes] = await Promise.all([
                api.get(queueEndpoint),
                api.get(backlogEndpoint),
                api.get(countsEndpoint)
            ]);

            setFallbackQueue(queueRes.data.data || []);
            setFallbackBacklog(backlogRes.data.count || 0);
            setFallbackBacklogValue(backlogRes.data.value || 0);
            setFallbackCounts(countsRes.data.counts || { RTS: 0, EHA: 0, SAL: 0 });
            setFallbackValues(countsRes.data.values || { RTS: 0, EHA: 0, SAL: 0 });
            setFallbackAvgTimes(countsRes.data.avgTimes || { RTS: 0.0, EHA: 0.0 }); // NOVO
            console.log("fetchDataApi: Dados de fallback atualizados.");

        } catch (error) {
            console.error("Falha ao buscar dados via API", error);
            setFallbackQueue([]);
            setFallbackBacklog(0);
            setFallbackBacklogValue(0);
            setFallbackCounts({ RTS: 0, EHA: 0, SAL: 0 });
            setFallbackValues({ RTS: 0, EHA: 0, SAL: 0 });
            setFallbackAvgTimes({ RTS: 0.0, EHA: 0.0 }); // NOVO
        } finally {
             console.log("fetchDataApi: Marcando initialDataLoaded como true.");
            setInitialDataLoaded(true);
        }
    }, [isGuest, initialDataLoaded]);

    // ... (useEffect syncTime, fetchDataApi, syncedTime inalterados) ...
	useEffect(() => {
        const syncTime = async () => { /* ... (inalterado) ... */
             try {
                const response = await api.get('/public/time');
                const serverTime = new Date(response.data.serverTime).getTime();
                const localTime = new Date().getTime();
                setTimeOffset(serverTime - localTime);
                 console.log("Tempo sincronizado.");
            } catch (error) {
                console.error("Falha ao sincronizar o tempo:", error);
                setTimeOffset(0);
            }
        };
        syncTime();
    }, []);

     useEffect(() => {
        if ((isConnected && !isGuest && wsQueue.length > 0 && !initialDataLoaded) ) {
             console.log("Dados iniciais recebidos via WebSocket.");
            setInitialDataLoaded(true);
        }

        if (!initialDataLoaded) {
             fetchDataApi();
        }

     }, [isConnected, isGuest, wsQueue, initialDataLoaded, fetchDataApi]); 

    useEffect(() => { /* ... (inalterado) ... */
        const interval = setInterval(() => {
            setSyncedTime(new Date().getTime() + timeOffset);
        }, 1000);
        return () => clearInterval(interval);
    }, [timeOffset]);

    // --- SELEÇÃO DE DADOS ATUALIZADA ---
    const useFallbackData = isGuest || !isConnected;
    const currentQueue = useFallbackData ? fallbackQueue : wsQueue;
    const currentBacklog = useFallbackData ? fallbackBacklog : wsBacklog; // Contagem backlog
    const currentBacklogValue = useFallbackData ? fallbackBacklogValue : wsBacklogValue; // Soma backlog
    const currentCounts = useFallbackData ? fallbackCounts : wsBufferCounts; // Contagem buffers
    const currentValues = useFallbackData ? fallbackValues : wsBufferValues; // Soma buffers
    const currentAvgTimes = useFallbackData ? fallbackAvgTimes : wsBufferAvgTimes; // Tempo médio buffers


    // ... (oldestDurations, filteredQueue, modais inalterados) ...
	const oldestDurations = useMemo(() => { /* ... (lógica inalterada, usa currentQueue) ... */
        const now = syncedTime;
        let rtsSeconds = 0;
        let ehaSeconds = 0;
        const oldestRTS = currentQueue.find(item => item.Buffer === 'RTS');
        const oldestEHA = currentQueue.find(item => item.Buffer === 'EHA');
        if (oldestRTS) {
            rtsSeconds = Math.max(0, Math.floor((now - new Date(oldestRTS.EntryTimestamp).getTime()) / 1000));
        }
        if (oldestEHA) {
            ehaSeconds = Math.max(0, Math.floor((now - new Date(oldestEHA.EntryTimestamp).getTime()) / 1000));
        }
        return {
            rts: { item: oldestRTS, duration: rtsSeconds },
            eha: { item: oldestEHA, duration: ehaSeconds }
        };
    }, [currentQueue, syncedTime]);

    const filteredQueue = useMemo(() => { /* ... (lógica inalterada, usa currentQueue) ... */
         if (filterBuffer === 'ALL') return currentQueue;
        return currentQueue.filter(item => item.Buffer === filterBuffer);
    }, [currentQueue, filterBuffer]);

    const openMoveModal = (item) => {/* ... */ setSelectedItem(item); setMoveModalOpen(true);};
    const closeMoveModal = () => {/* ... */ setSelectedItem(null); setMoveModalOpen(false);};
    const handleSuccess = useCallback(() => {/* ... */ console.log("Ação concluída, aguardando WS.");}, []);

    if (!initialDataLoaded) {
        return <p className="loading-message">A carregar dados...</p>;
    }

    return (
        <div className="app-container dashboard-container">
            {/* ... (mensagem de conexão, header) ... */}
			{!isGuest && !isConnected && (
                <p style={{ color: 'orange', textAlign: 'center', marginBottom: '1rem' }}>
                    Sem conexão em tempo real. Tentando reconectar...
                </p>
            )}

            <header className="dashboard-header">
               <div><h1>FIFO</h1><p>Sistema de Controle Logístico</p></div>
               <div className="user-profile"><span>{user?.username}</span>{!isGuest && (<button onClick={() => setChangePasswordModalOpen(true)} className="change-password-button">ALTERAR SENHA</button>)}<button onClick={logout} className="logout-button">SAIR</button></div>
            </header>

            <main>
                {/* --- SEÇÃO DE MÉTRICAS ATUALIZADA (6 CARDS) --- */}
                <section className="metrics-grid">
                    {/* Linha 1 */}
                    <div className="metric-card">
                        <span className="metric-value">{currentBacklogValue}</span>
                        <span className="metric-label">Back-Log Total (Produtos)</span>
                        <span className="metric-label" style={{color: 'var(--color-primary)', marginTop: '4px'}}>({currentBacklog} Itens)</span>
                    </div>
                     <div className="metric-card buffer-card"> {/* Card Quantidades */}
                        <div className="buffer-count"><span>RTS:</span><span>{currentCounts.RTS}</span></div>
                        <div className="buffer-count"><span>EHA:</span><span>{currentCounts.EHA}</span></div>
                        <div className="buffer-count"><span>SALVADOS:</span><span>{currentCounts.SAL}</span></div>
                    </div>
                    <div className="metric-card buffer-card"> {/* Card Backlog Buffers */}
                        <div className="buffer-count"><span>RTS:</span><span>{currentValues.RTS}</span></div>
                        <div className="buffer-count"><span>EHA:</span><span>{currentValues.EHA}</span></div>
                        <div className="buffer-count" style={{opacity: 0.5}}><span>SALVADOS:</span><span>{currentValues.SAL}</span></div> {/* SAL sempre 0 */}
                    </div>

                    {/* Linha 2 */}
                    <div className="metric-card"> {/* Card Maior Tempo RTS */}
                        <span className="metric-value">{formatDuration(oldestDurations.rts.duration)}</span>
                        <span className="metric-sub-label">{oldestDurations.rts.item ? oldestDurations.rts.item.TrackingID : '---'}</span>
                        <span className="metric-label">Maior Tempo RTS</span>
                    </div>
                    <div className="metric-card"> {/* Card Maior Tempo EHA */}
                        <span className="metric-value">{formatDuration(oldestDurations.eha.duration)}</span>
                        <span className="metric-sub-label">{oldestDurations.eha.item ? oldestDurations.eha.item.TrackingID : '---'}</span>
                        <span className="metric-label">Maior Tempo EHA</span>
                    </div>
                     <div className="metric-card buffer-card"> {/* Card Tempo Médio */}
                        <div className="buffer-count"><span>RTS Médio:</span><span>{formatDuration(currentAvgTimes.RTS)}</span></div>
                        <div className="buffer-count"><span>EHA Médio:</span><span>{formatDuration(currentAvgTimes.EHA)}</span></div>
                         {/* Placeholder ou espaço vazio */}
                         <div className="buffer-count" style={{visibility: 'hidden'}}><span>&nbsp;</span><span>&nbsp;</span></div>
                    </div>
                </section>
                {/* --- FIM DA SEÇÃO DE MÉTRICAS --- */}

                {/* ... (Filtro) ... */}
				<section className="filter-controls">
                  <label htmlFor="buffer-filter" className="filter-label">Filtrar Buffer:</label>
                  <select id="buffer-filter" className="dashboard-filter-select" value={filterBuffer} onChange={(e) => setFilterBuffer(e.target.value)}>
                    <option value="ALL">TODOS ({currentCounts.RTS + currentCounts.EHA + currentCounts.SAL})</option> {/* Soma total */}
                    <option value="RTS">RTS ({currentCounts.RTS})</option>
                    <option value="EHA">EHA ({currentCounts.EHA})</option>
                    <option value="SAL">SALVADOS ({currentCounts.SAL})</option> {/* Adiciona filtro SAL */}
                  </select>
                </section>

                <section className="fifo-list">
                    {/* --- HEADER DA LISTA COM PERFIL OPCIONAL --- */}
                    <header className={`fifo-list-header ${!isGuest ? 'with-actions' : ''}`}>
                       <span>ID</span>
                       <span>PERFIL</span>
                       <span>BUFFER</span>
                       <span>RUA</span>
                       <span>DURAÇÃO</span>
                       {!isGuest && <span>AÇÕES</span>}
                    </header>
                    <div className="fifo-list-body">
                        {filteredQueue.length > 0 ? filteredQueue.map(item => {
                            const entryTimestamp = new Date(item.EntryTimestamp).getTime();
                            const durationSeconds = Math.max(0, Math.floor((syncedTime - entryTimestamp) / 1000));
                            return (
                                /* --- ITEM DA LISTA COM PERFIL CONDICIONAL --- */
                                <div className={`fifo-list-item ${!isGuest ? 'with-actions' : ''}`} key={item.ID}>
                                    <span>{item.TrackingID}</span>
                                    {/* Exibe perfil ou '---' se for N/A */}
                                    <span>{item.Profile !== 'N/A' ? item.Profile : '---'}</span>
                                    <span>{item.Buffer}</span>
                                    <span>{item.Rua}</span>
                                    <span>{formatDuration(durationSeconds)}</span>
                                    {!isGuest && (
                                        <div className="action-buttons-cell">
                                            {hasPermission('MOVE_PACKAGE') && (<button onClick={() => openMoveModal(item)} className="move-btn">Mover</button>)}
                                        </div>
                                    )}
                                </div>
                            );
                        }) : (
                           <p className="empty-queue-message">{filterBuffer === 'ALL' ? 'A fila está vazia.' : `Nenhum item no buffer ${filterBuffer}.`}</p>
                        )}
                    </div>
                </section>

                 {/* ... (Botões de Ação e Navegação) ... */}
				 {!isGuest && hasPermission('MANAGE_FIFO') && (<section className="actions-grid"><button className="action-button entry" onClick={() => setEntryModalOpen(true)}>ENTRADA</button><button className="action-button exit" onClick={() => setExitModalOpen(true)}>SAÍDA</button></section>)}
                 <div className="admin-nav-buttons">{!isGuest && hasPermission('GENERATE_QR_CODES') && (<button onClick={() => navigate('/qrcode-generator')} className="admin-nav-button">GERAR QR CODES</button>)}{!isGuest && hasPermission('VIEW_LOGS') && (<button onClick={() => navigate('/logs')} className="admin-nav-button">VER LOGS DE ATIVIDADE</button>)}{!isGuest && hasPermission('VIEW_USERS') && (<button onClick={() => navigate('/management')} className="admin-nav-button">PAINEL DE GESTÃO</button>)}</div>
            </main>

            {/* ... (Modais) ... */}
			<ChangePasswordModal isOpen={isChangePasswordModalOpen} onClose={() => setChangePasswordModalOpen(false)} />
            <EntryModal isOpen={isEntryModalOpen} onClose={() => setEntryModalOpen(false)} onSuccess={handleSuccess} />
            <ExitModal isOpen={isExitModalOpen} onClose={() => setExitModalOpen(false)} onSuccess={handleSuccess} availableIDs={currentQueue.map(item => item.TrackingID)} />
            <MoveItemModal isOpen={isMoveModalOpen} onClose={closeMoveModal} onSuccess={handleSuccess} item={selectedItem} />
        </div>
    );
}

export default DashboardPage;