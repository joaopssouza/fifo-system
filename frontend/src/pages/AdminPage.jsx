// src/pages/AdminPage.jsx
import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';
import { useAuth } from '../context/AuthContext';
import { useWebSocket } from '../context/WebSocketContext';

// --- PASSO 1: Definir a lista de setores disponíveis ---
const SECTORS = ["ADMINISTRAÇÃO", "FIFO", "LIDERANÇA", "OPERAÇÕES", "GERAL"];

// --- PASSO 2: Atualizar o Modal de Criação ---
const CreateUserModal = ({ isOpen, onClose, onSuccess, roles, actingUserRole }) => {
    const assignableRoles = useMemo(() => {
        if (actingUserRole === 'admin') {
            return roles; 
        }
        return roles.filter(role => role.Name !== 'admin' && role.Name !== 'leader');
    }, [roles, actingUserRole]);

    const [username, setUsername] = useState('');
    const [fullName, setFullName] = useState('');
    const [password, setPassword] = useState('');
    const [role, setRole] = useState(assignableRoles[0]?.Name || 'fifo');
    const [sector, setSector] = useState(SECTORS[0]); // Define o setor inicial como o primeiro da lista
    const [error, setError] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        try {
            await api.post('/api/management/users', { username, fullName, password, role, sector });
            onSuccess();
            onClose();
        } catch (err) {
            setError(err.response?.data?.error || "Falha ao criar utilizador.");
        }
    };

    if (!isOpen) return null;

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal-content" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>Criar Novo Utilizador</h2>
                    <button className="modal-close-button" onClick={onClose}>&times;</button>
                </div>
                <form onSubmit={handleSubmit}>
                    <label>Nome Completo</label>
                    <input type="text" value={fullName} onChange={e => setFullName(e.target.value)} required />

                    <label>Nome de Utilizador</label>
                    <input type="text" value={username} onChange={e => setUsername(e.target.value)} required />
                    <label>Senha</label>
                    <input type="password" value={password} onChange={e => setPassword(e.target.value)} required minLength="6" />
                    <label>Papel</label>
                    <select value={role} onChange={e => setRole(e.target.value)} required>
                        {assignableRoles.map(r => (
                            <option key={r.ID} value={r.Name}>{r.Name.toUpperCase()}</option>
                        ))}
                    </select>
                    {/* O campo de texto do setor foi substituído por um <select> */}
                    <label>Setor</label>
                    <select value={sector} onChange={e => setSector(e.target.value)} required>
                        {SECTORS.map(s => (
                            <option key={s} value={s}>{s}</option>
                        ))}
                    </select>
                    <button type="submit" className="modal-submit-button blue">Criar Utilizador</button>
                    {error && <p className="error-message">{error}</p>}
                </form>
            </div>
        </div>
    );
};

// --- PASSO 3: Atualizar o Modal de Edição ---
const EditUserModal = ({ isOpen, onClose, onSuccess, user, roles }) => {
    const [fullName, setFullName] = useState('');
    const [roleId, setRoleId] = useState('');
    const [sector, setSector] = useState('');
    const [error, setError] = useState('');

    useEffect(() => {
        if (user) {
            setFullName(user.FullName);
            setRoleId(user.RoleID);
            setSector(user.Sector);
        }
    }, [user]);

    const handleUpdate = async (e) => {
        e.preventDefault();
        setError('');
        try {
            await api.put(`/api/management/users/${user.ID}`, {fullName, roleId: parseInt(roleId, 10), sector });
            onSuccess();
            onClose();
        } catch (err) {
            setError(err.response?.data?.error || "Falha ao atualizar utilizador.");
        }
    };

    if (!isOpen || !user) return null;

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal-content" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2>Editar Utilizador: {user.Username.toUpperCase()}</h2>
                    <button className="modal-close-button" onClick={onClose}>&times;</button>
                </div>
                <form onSubmit={handleUpdate}>
                    <label>Nome Completo</label>
                    <input type="text" value={fullName} onChange={e => setFullName(e.target.value)} required />

                    <label>Papel</label>
                    <select value={roleId} onChange={e => setRoleId(e.target.value)}>
                        {roles
                            .filter(role => role.Name !== 'admin')
                            .map(role => (
                                <option key={role.ID} value={role.ID}>{role.Name.toUpperCase()}</option>
                            ))
                        }
                    </select>
                    {/* O campo de texto do setor foi substituído por um <select> */}
                    <label>Setor</label>
                    <select value={sector} onChange={e => setSector(e.target.value)} required>
                        {SECTORS.map(s => (
                            <option key={s} value={s}>{s}</option>
                        ))}
                    </select>
                    <button type="submit" className="modal-submit-button blue">Salvar Alterações</button>
                    {error && <p className="error-message">{error}</p>}
                </form>
            </div>
        </div>
    );
};

const ResetPasswordModal = ({ isOpen, onClose, user }) => { const [newPassword, setNewPassword] = useState(''); const [error, setError] = useState(''); const handleReset = async (e) => { e.preventDefault(); setError(''); try { await api.put(`/api/management/users/${user.ID}/reset-password`, { newPassword }); onClose(); } catch (err) { setError(err.response?.data?.error || "Falha ao redefinir a senha."); } }; if (!isOpen || !user) return null; return (<div className="modal-overlay" onClick={onClose}><div className="modal-content" onClick={e => e.stopPropagation()}><div className="modal-header"><h2>Redefinir Senha de: {user.Username.toUpperCase()}</h2><button className="modal-close-button" onClick={onClose}>&times;</button></div><form onSubmit={handleReset}><label>Nova Senha</label><input type="password" value={newPassword} onChange={e => setNewPassword(e.target.value)} required minLength="6" /><button type="submit" className="modal-submit-button red">Redefinir Senha</button>{error && <p className="error-message">{error}</p>}</form></div></div>); };

function AdminPage() {
    const { user: actingUser, hasPermission } = useAuth();
    const { onlineUsers } = useWebSocket();
    const navigate = useNavigate();
    const [users, setUsers] = useState([]);
    const [roles, setRoles] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedUser, setSelectedUser] = useState(null);
    const [isCreateModalOpen, setCreateModalOpen] = useState(false);
    const [isEditModalOpen, setEditModalOpen] = useState(false);
    const [isResetModalOpen, setResetModalOpen] = useState(false);

    const fetchAdminData = useCallback(async () => {
        setLoading(true);
        try {
            const [usersRes, rolesRes] = await Promise.all([
                api.get('/api/management/users'),
                api.get('/api/management/roles')
            ]);
            setUsers(usersRes.data.data || []);
            setRoles(rolesRes.data.data || []);
        } catch (error) { console.error("Falha ao buscar dados de administração", error); } 
        finally { setLoading(false); }
    }, []);

    useEffect(() => { fetchAdminData(); }, [fetchAdminData]);
    
        const canPerformActionsOn = (targetUser) => {
        // Regra 1: Ninguém pode realizar ações em si mesmo.
        if (actingUser.username === targetUser.Username) {
            return false;
        }

        // Regra 2: Ninguém pode realizar ações em um 'admin'.
        if (targetUser.Role.Name === 'admin') {
            return false;
        }

        // Regra 3: Se o utilizador logado for um 'admin', ele pode realizar ações em qualquer um (que não seja outro admin, já verificado acima).
        if (actingUser.role === 'admin') {
            return true;
        }

        // Regra 4: Se o utilizador logado for um 'leader', ele NÃO PODE realizar ações em outros 'leaders'.
        if (actingUser.role === 'leader') {
            if (targetUser.Role.Name === 'leader') {
                return false;
            }
            return true; // Mas pode realizar ações em papéis inferiores (ex: 'fifo').
        }

        // Por defeito, outros papéis não podem realizar ações.
        return false;
    };

    const openEditModal = (user) => { setSelectedUser(user); setEditModalOpen(true); };
    const openResetModal = (user) => { setSelectedUser(user); setResetModalOpen(true); };

    return (
        <div className="app-container admin-container">
            <header className="admin-header">
                <h1>Painel de Administração</h1>
                <button onClick={() => navigate('/')} className="back-button">Voltar ao Dashboard</button>
            </header>
            <section className="admin-section">
                <h2>Utilizadores Online ({onlineUsers.length})</h2>
                <div className="table-container">
                     <table className="admin-table">
                        <thead><tr><th>Nome Completo</th><th>Utilizador</th><th>Papel</th><th>Setor</th><th>ONLINE</th></tr></thead>
                        <tbody>
                            {onlineUsers.length > 0 ? (
                                onlineUsers.map((user, index) => (<tr key={`online-${user.id}-${index}`}><td>{user.fullName.toUpperCase()}</td><td>{user.username.toUpperCase()}</td><td>{user.role.toUpperCase()}</td><td>{user.sector.toUpperCase()}</td><td style={{ textAlign: 'center' }}>🟢</td></tr>))
                            ) : ( <tr><td colSpan="5">Nenhum utilizador online.</td></tr> )}
                        </tbody>
                    </table>
                </div>
            </section>
            <section className="admin-section">
                <div className="section-header">
                    <h2>Todos os Utilizadores ({users.length})</h2>
                    {hasPermission('CREATE_USER') && (<button onClick={() => setCreateModalOpen(true)} className="action-button entry">Criar Utilizador</button>)}
                </div>
                <div className="table-container">
                    <table className="admin-table">
                         <thead><tr><th>Nome Completo</th><th>Utilizador</th><th>Papel</th><th>Setor</th><th>Ações</th></tr></thead>
                        <tbody>
                            {loading ? (<tr><td colSpan="4">A carregar...</td></tr>) 
                            : users.map(user => (
                                <tr key={user.ID}>
                                    <td>{user.FullName.toUpperCase()}</td>
                                    <td>{user.Username.toUpperCase()}</td>
                                    <td>{user.Role.Name.toUpperCase()}</td>
                                    <td>{user.Sector.toUpperCase()}</td>
                                    <td>
                                        {/* --- LÓGICA DE EXIBIÇÃO SIMPLIFICADA E CORRIGIDA --- */}
                                        {canPerformActionsOn(user) && (
                                            <div className="action-buttons-cell">
                                                {hasPermission('EDIT_USER') && (<button onClick={() => openEditModal(user)} className="edit-btn">Editar</button>)}
                                                {hasPermission('RESET_PASSWORD') && (<button onClick={() => openResetModal(user)} className="reset-btn">Redefinir Senha</button>)}
                                            </div>
                                        )}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </section>
            <CreateUserModal 
                isOpen={isCreateModalOpen} 
                onClose={() => setCreateModalOpen(false)} 
                onSuccess={fetchAdminData} 
                roles={roles}
                actingUserRole={actingUser?.role}
            />
            <EditUserModal 
                isOpen={isEditModalOpen} 
                onClose={() => setEditModalOpen(false)} 
                onSuccess={fetchAdminData} 
                user={selectedUser} 
                roles={roles} 
            />
            <ResetPasswordModal 
                isOpen={isResetModalOpen} 
                onClose={() => setResetModalOpen(false)} 
                user={selectedUser} 
            />
        </div>
    );
}

export default AdminPage;

