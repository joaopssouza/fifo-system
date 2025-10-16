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

    const [filters, setFilters] = useState({
        username: '',
        fullname: '',
        action: '',
        startDate: '',
        endDate: ''
    });

    const fetchLogs = useCallback(async () => {
        setLoading(true);
        try {
            const params = new URLSearchParams();
            if (filters.username) params.append('username', filters.username);
            if (filters.fullname) params.append('fullname', filters.fullname);
            if (filters.action) params.append('action', filters.action);
            if (filters.startDate && filters.endDate) {
                params.append('startDate', filters.startDate);
                params.append('endDate', filters.endDate);
            }

            const response = await api.get(`/api/management/logs?${params.toString()}`);
            setLogs(response.data.data || []);
        } catch (err) {
            setError('Falha ao carregar logs. Você tem permissão de administrador?');
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, [filters]);

    useEffect(() => {
        fetchLogs();
    }, [fetchLogs]);

    const handleFilterChange = (e) => {
        const { name, value } = e.target;
        setFilters(prev => ({ ...prev, [name]: value }));
    };

    const clearFilters = () => {
        setFilters({ fullname: '', username: '', action: '', startDate: '', endDate: '' });
    };

    return (
        <div className="app-container logs-container">
            <header className="logs-header">
                <h1>Logs de Atividade</h1>
                <button onClick={() => navigate('/')} className="back-button">Voltar ao Dashboard</button>
            </header>

            {/* --- PAINEL DE FILTROS MODIFICADO --- */}
            <div className="filters-panel">
                <div className="input-with-label">
                    <label htmlFor="username">Filtrar por utilizador</label>
                    <input
                        id="username"
                        type="text"
                        name="username"
                        placeholder="Digite o nome de utilizador..."
                        value={filters.username}
                        onChange={handleFilterChange}
                    />
                </div>
                <div className="input-with-label">
                    <label htmlFor="fullname">Filtrar por nome completo</label>
                    <input
                        id="fullname"
                        type="text"
                        name="fullname"
                        placeholder="Digite o nome completo..."
                        value={filters.fullname}
                        onChange={handleFilterChange}
                    />
                </div>
                <div className="input-with-label">
                    <label htmlFor="action">Filtrar por ação</label>
                    <select id="action" name="action" value={filters.action} onChange={handleFilterChange}>
                        <option value="">Todas as Ações</option>
                        <option value="ENTRADA">Entrada</option>
                        <option value="SAIDA">Saída</option>
                        <option value="MOVIMENTACAO">Movimentação</option>
                    </select>
                </div>
                <div className="input-with-label">
                    <label htmlFor="startDate">Data de Início</label>
                    <input
                        id="startDate"
                        type="date"
                        name="startDate"
                        value={filters.startDate}
                        onChange={handleFilterChange}
                    />
                </div>
                <div className="input-with-label">
                    <label htmlFor="endDate">Data de Fim</label>
                    <input
                        id="endDate"
                        type="date"
                        name="endDate"
                        value={filters.endDate}
                        onChange={handleFilterChange}
                    />
                </div>
                <button onClick={clearFilters} className="clear-filters-button">Limpar Filtros</button>
            </div>
            {/* --- FIM DA MODIFICAÇÃO --- */}


            <div className="table-container">
                <table className="logs-table">
                    <thead>
                        <tr>
                            <th>Data/Hora</th>
                            <th>Nome Completo</th>
                            <th>Utilizador</th>
                            <th>Ação</th>
                            <th>Detalhes</th>
                        </tr>
                    </thead>
                    <tbody>
                        {loading ? (
                            <tr><td colSpan="5">A carregar...</td></tr>
                        ) : logs.length > 0 ? (
                            logs.map(log => (
                                <tr key={log.ID}>
                                    <td>{formatTimestamp(log.CreatedAt)}</td>
                                    <td>{log.UserFullname}</td>
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
                                <td colSpan="5">Nenhum registo de atividade encontrado para os filtros selecionados.</td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
}

export default LogsPage;