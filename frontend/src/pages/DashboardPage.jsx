// src/pages/DashboardPage.jsx
import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useWebSocket } from '../context/WebSocketContext';
import api from '../services/api';

// ... (Modal imports and formatDuration) ...
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
    const { user, logout, hasPermission, isGuest } = useAuth();
    const { wsQueue, wsBacklog, wsBufferCounts, isConnected } = useWebSocket();
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

    // Estados de fallback (usados SE WebSocket não conectado OU se for convidado)
    const [fallbackQueue, setFallbackQueue] = useState([]);
    const [fallbackBacklog, setFallbackBacklog] = useState(0);
    const [fallbackCounts, setFallbackCounts] = useState({ RTS: 0, EHA: 0, SAL: 0 });

    // Busca dados via API (usado na carga inicial e como fallback)
    const fetchDataApi = useCallback(async () => {
        // Só busca se os dados iniciais ainda não foram carregados
        if (initialDataLoaded) {
             console.log("fetchDataApi: Skipping, initial data already loaded.");
             return;
        }
        console.log("fetchDataApi: Buscando dados via API...");
        // Define loading inicial aqui, antes da chamada
        // Usamos um estado separado para não confundir com o estado do WS
        // (Reintroduzindo isLoadingApi para clareza)
        // const [isLoadingApi, setIsLoadingApi] = useState(!initialDataLoaded); // Inicia true se não carregado
        // setIsLoadingApi(true); // Removido, estado inicial já é true se !initialDataLoaded

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
            setFallbackCounts(countsRes.data || { RTS: 0, EHA: 0, SAL: 0 });
            console.log("fetchDataApi: Dados de fallback atualizados.");

        } catch (error) {
            console.error("Falha ao buscar dados via API", error);
            setFallbackQueue([]);
            setFallbackBacklog(0);
            setFallbackCounts({ RTS: 0, EHA: 0, SAL: 0 });
        } finally {
            // Marca como carregado independentemente de sucesso ou falha da API
             console.log("fetchDataApi: Marcando initialDataLoaded como true.");
            setInitialDataLoaded(true);
            // setIsLoadingApi(false); // Removido
        }
    }, [isGuest, initialDataLoaded]); // Depende de isGuest e initialDataLoaded

    // Efeito para sincronizar tempo (executa uma vez)
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

     // Efeito para buscar dados iniciais via API (se necessário)
     useEffect(() => {
        // Se já temos dados da fila vindos do WS E ainda não marcamos como carregado, marca agora
        // OU se é convidado e ainda não marcamos como carregado (convidado não receberá dados WS)
        if ((isConnected && !isGuest && wsQueue.length > 0 && !initialDataLoaded) ) {
             console.log("Dados iniciais recebidos via WebSocket.");
            setInitialDataLoaded(true);
        }

        // Busca via API se ainda não tivermos carregado os dados iniciais.
        // Isso cobre o caso do convidado e o fallback para user autenticado.
        if (!initialDataLoaded) {
             fetchDataApi();
        }

     }, [isConnected, isGuest, wsQueue, initialDataLoaded, fetchDataApi]); // Observa a chegada de dados do WS e estado de loading


    // Efeito para atualizar a hora sincronizada (inalterado)
    useEffect(() => { /* ... (inalterado) ... */
        const interval = setInterval(() => {
            setSyncedTime(new Date().getTime() + timeOffset);
        }, 1000);
        return () => clearInterval(interval);
    }, [timeOffset]);


    // --- Seleção de Dados ---
    // Se for convidado, usa SEMPRE o fallback.
    // Se for autenticado, usa WS se conectado, senão usa fallback.
    const useFallbackData = isGuest || !isConnected;
    const currentQueue = useFallbackData ? fallbackQueue : wsQueue;
    const currentBacklog = useFallbackData ? fallbackBacklog : wsBacklog;
    const currentCounts = useFallbackData ? fallbackCounts : wsBufferCounts;


    // --- Cálculos usam os dados correntes ---
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

    // Funções de controle dos modais (inalteradas)
    const openMoveModal = (item) => {/* ... */ setSelectedItem(item); setMoveModalOpen(true);};
    const closeMoveModal = () => {/* ... */ setSelectedItem(null); setMoveModalOpen(false);};
    const handleSuccess = useCallback(() => {/* ... */ console.log("Ação concluída, aguardando WS.");}, []);

    // --- Lógica de Loading SIMPLIFICADA ---
    // Mostra loading apenas se os dados iniciais AINDA não foram marcados como carregados
    if (!initialDataLoaded) {
        return <p className="loading-message">A carregar dados...</p>;
    }

    return (
        <div className="app-container dashboard-container">
            {/* --- AJUSTE NA MENSAGEM DE CONEXÃO --- */}
            {/* Mostra APENAS se NÃO for convidado E NÃO estiver conectado */}
            {!isGuest && !isConnected && (
                <p style={{ color: 'orange', textAlign: 'center', marginBottom: '1rem' }}>
                    Sem conexão em tempo real. Tentando reconectar...
                </p>
            )}
            {/* --- FIM DO AJUSTE --- */}


            <header className="dashboard-header">
               {/* ... (header inalterado) ... */}
               <div><h1>FIFO</h1><p>Sistema de Controle Logístico</p></div>
               <div className="user-profile"><span>{user?.username}</span>{!isGuest && (<button onClick={() => setChangePasswordModalOpen(true)} className="change-password-button">ALTERAR SENHA</button>)}<button onClick={logout} className="logout-button">SAIR</button></div>
            </header>

            <main>
                {/* --- SEÇÕES USAM DADOS CORRENTES (Determinados pela lógica corrigida) --- */}
                <section className="metrics-grid">
                    <div className="metric-card"><span className="metric-value">{currentBacklog}</span><span className="metric-label">Itens na Fila</span></div>
                    <div className="metric-card buffer-card"><div className="buffer-count"><span>RTS:</span><span>{currentCounts.RTS}</span></div><div className="buffer-count"><span>EHA:</span><span>{currentCounts.EHA}</span></div><div className="buffer-count"><span>SALVADOS:</span><span>{currentCounts.SAL}</span></div></div>
                    <div className="metric-card"><span className="metric-value">{formatDuration(oldestDurations.rts.duration)}</span><span className="metric-sub-label">{oldestDurations.rts.item ? oldestDurations.rts.item.TrackingID : '---'}</span><span className="metric-label">Maior Tempo RTS</span></div>
                    <div className="metric-card"><span className="metric-value">{formatDuration(oldestDurations.eha.duration)}</span><span className="metric-sub-label">{oldestDurations.eha.item ? oldestDurations.eha.item.TrackingID : '---'}</span><span className="metric-label">Maior Tempo EHA</span></div>
                </section>

                <section className="filter-controls">
                  <label htmlFor="buffer-filter" className="filter-label">Filtrar Buffer:</label>
                  <select id="buffer-filter" className="dashboard-filter-select" value={filterBuffer} onChange={(e) => setFilterBuffer(e.target.value)}>
                    <option value="ALL">TODOS ({currentQueue.length})</option>
                    <option value="RTS">RTS ({currentCounts.RTS})</option>
                    <option value="EHA">EHA ({currentCounts.EHA})</option>
                  </select>
                </section>

                <section className="fifo-list">
                    <header className={`fifo-list-header ${!isGuest ? 'with-actions' : ''}`}>
                       {/* ... (header inalterado) ... */}
                       <span>ID</span><span>BUFFER</span><span>RUA</span><span>DURAÇÃO</span>{!isGuest && <span>AÇÕES</span>}
                    </header>
                    <div className="fifo-list-body">
                         {/* Usa filteredQueue (que depende de currentQueue) */}
                        {filteredQueue.length > 0 ? filteredQueue.map(item => {
                            const entryTimestamp = new Date(item.EntryTimestamp).getTime();
                            const durationSeconds = Math.max(0, Math.floor((syncedTime - entryTimestamp) / 1000));
                            return (
                                <div className={`fifo-list-item ${!isGuest ? 'with-actions' : ''}`} key={item.ID}>
                                    {/* ... (renderização do item inalterada) ... */}
                                    <span>{item.TrackingID}</span><span>{item.Buffer}</span><span>{item.Rua}</span><span>{formatDuration(durationSeconds)}</span>{!isGuest && (<div className="action-buttons-cell">{hasPermission('MOVE_PACKAGE') && (<button onClick={() => openMoveModal(item)} className="move-btn">Mover</button>)}</div>)}
                                </div>
                            );
                        }) : (
                           // ... (mensagem fila vazia inalterada) ...
                           <p className="empty-queue-message">{filterBuffer === 'ALL' ? 'A fila está vazia.' : `Nenhum item no buffer ${filterBuffer}.`}</p>
                        )}
                    </div>
                </section>

                {/* ... (Restante do JSX inalterado) ... */}
                 {!isGuest && hasPermission('MANAGE_FIFO') && (<section className="actions-grid"><button className="action-button entry" onClick={() => setEntryModalOpen(true)}>ENTRADA</button><button className="action-button exit" onClick={() => setExitModalOpen(true)}>SAÍDA</button></section>)}
                 <div className="admin-nav-buttons">{!isGuest && hasPermission('GENERATE_QR_CODES') && (<button onClick={() => navigate('/qrcode-generator')} className="admin-nav-button">GERAR QR CODES</button>)}{!isGuest && hasPermission('VIEW_LOGS') && (<button onClick={() => navigate('/logs')} className="admin-nav-button">VER LOGS DE ATIVIDADE</button>)}{!isGuest && hasPermission('VIEW_USERS') && (<button onClick={() => navigate('/management')} className="admin-nav-button">PAINEL DE GESTÃO</button>)}</div>
            </main>

             {/* Modais usam currentQueue para availableIDs */}
            <ChangePasswordModal isOpen={isChangePasswordModalOpen} onClose={() => setChangePasswordModalOpen(false)} />
            <EntryModal isOpen={isEntryModalOpen} onClose={() => setEntryModalOpen(false)} onSuccess={handleSuccess} />
            <ExitModal isOpen={isExitModalOpen} onClose={() => setExitModalOpen(false)} onSuccess={handleSuccess} availableIDs={currentQueue.map(item => item.TrackingID)} />
            <MoveItemModal isOpen={isMoveModalOpen} onClose={closeMoveModal} onSuccess={handleSuccess} item={selectedItem} />
        </div>
    );
}

export default DashboardPage;