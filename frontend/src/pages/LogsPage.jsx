// src/pages/LogsPage.jsx
import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';

const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleString('pt-BR');
};

function LogsPage() {
    const [logs, setLogs] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const navigate = useNavigate();

    // --- NOVO: Estado para os filtros ---
    const [filters, setFilters] = useState({
        username: '',
        action: '', // 'ENTRADA', 'SAIDA', ou '' para todos
        startDate: '',
        endDate: ''
    });

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            // Constrói os parâmetros da URL a partir do estado dos filtros
            const params = new URLSearchParams();
            if (filters.username) params.append('username', filters.username);
            if (filters.action) params.append('action', filters.action);
            if (filters.startDate && filters.endDate) {
                params.append('startDate', filters.startDate);
                params.append('endDate', filters.endDate);
            }

            const response = await api.get(`/api/admin/logs?${params.toString()}`);
            setLogs(response.data.data || []);
        } catch (err) {
            setError('Falha ao carregar logs. Você tem permissão de administrador?');
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, [filters]); // A função será recriada se os filtros mudarem

    useEffect(() => {
        fetchLogs();
    }, [fetchLogs]); // Executa a busca inicial e sempre que a função fetchLogs for atualizada

    const handleFilterChange = (e) => {
        const { name, value } = e.target;
        setFilters(prev => ({ ...prev, [name]: value }));
    };

    const clearFilters = () => {
        setFilters({ username: '', action: '', startDate: '', endDate: '' });
    };

    return (
        <div className="app-container logs-container">
            <header className="logs-header">
                <h1>Logs de Atividade</h1>
                <button onClick={() => navigate('/')} className="back-button">Voltar ao Dashboard</button>
            </header>

            {/* --- NOVO: Painel de Filtros --- */}
            <div className="filters-panel">
                <input
                    type="text"
                    name="username"
                    placeholder="Filtrar por utilizador..."
                    value={filters.username}
                    onChange={handleFilterChange}
                />
                <select name="action" value={filters.action} onChange={handleFilterChange}>
                    <option value="">Todas as Ações</option>
                    <option value="ENTRADA">Entrada</option>
                    <option value="SAIDA">Saída</option>
                </select>
                <input
                    type="date"
                    name="startDate"
                    value={filters.startDate}
                    onChange={handleFilterChange}
                />
                <input
                    type="date"
                    name="endDate"
                    value={filters.endDate}
                    onChange={handleFilterChange}
                />
                <button onClick={clearFilters} className="clear-filters-button">Limpar Filtros</button>
            </div>

            <div className="table-container">
                <table className="logs-table">
                    {/* ... (cabeçalho da tabela sem alterações) ... */}
                    <thead>
                        <tr>
                            <th>Data/Hora</th>
                            <th>Utilizador</th>
                            <th>Ação</th>
                            <th>Detalhes</th>
                        </tr>
                    </thead>
                    <tbody>
                        {loading ? (
                            <tr><td colSpan="4">A carregar...</td></tr>
                        ) : logs.length > 0 ? (
                            logs.map(log => (
                                <tr key={log.ID}>
                                    <td>{formatTimestamp(log.CreatedAt)}</td>
                                    <td>{log.Username}</td>
                                    <td>
                                        <span className={`action-tag ${log.Action.toLowerCase()}`}>
                                            {log.Action}
                                        </span>
                                    </td>
                                    <td>{log.Details}</td>
                                </tr>
                            ))
                        ) : (
                            <tr>
                                <td colSpan="4">Nenhum registo de atividade encontrado para os filtros selecionados.</td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
}

export default LogsPage;
