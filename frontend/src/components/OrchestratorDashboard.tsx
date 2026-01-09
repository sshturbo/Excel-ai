import React, { useState, useEffect } from 'react';
import { Card } from './ui/card';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, BarChart, Bar } from 'recharts';
import './OrchestratorDashboard.css';

interface OrchestratorStats {
  totalTasks: number;
  successTasks: number;
  failedTasks: number;
  activeWorkers: number;
  avgTaskTime: string;
  successRate: number;
  isRunning: boolean;
}

interface HealthStatus {
  isHealthy: boolean;
  workersActive: number;
  totalTasks: number;
  tasksPending: number;
  lastCheck: string;
  issues: string[];
}

interface DashboardProps {
  onClose?: () => void;
}

export const OrchestratorDashboard: React.FC<DashboardProps> = ({ onClose }) => {
  const [stats, setStats] = useState<OrchestratorStats | null>(null);
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [history, setHistory] = useState<any[]>([]);
  const [isMonitoring, setIsMonitoring] = useState(false);

  // Fetch stats e health
  const fetchData = async () => {
    try {
      const s = await (window as any).go.main.GetOrchestratorStats();
      const h = await (window as any).go.main.OrchestratorHealthCheck();
      
      setStats(s);
      setHealth(h);
      
      // Atualizar histórico
      setHistory(prev => [...prev.slice(-59), { // Manter últimos 60 pontos
        timestamp: new Date().toLocaleTimeString(),
        successRate: s.successRate,
        activeWorkers: s.activeWorkers,
        totalTasks: s.totalTasks,
        successTasks: s.successTasks,
        failedTasks: s.failedTasks,
        avgTaskTime: s.avgTaskTime
      }]);
    } catch (error) {
      console.error('Erro ao buscar dados:', error);
    }
  };

  // Toggle monitoring
  const toggleMonitoring = async () => {
    if (isMonitoring) {
      setIsMonitoring(false);
    } else {
      setIsMonitoring(true);
      await fetchData();
    }
  };

  // Clear cache
  const clearCache = async () => {
    try {
      await (window as any).go.main.ClearOrchestratorCache();
      alert('Cache limpo com sucesso!');
      await fetchData();
    } catch (error) {
      alert('Erro ao limpar cache: ' + error);
    }
  };

  // Trigger recovery
  const triggerRecovery = async () => {
    try {
      await (window as any).go.main.TriggerOrchestratorRecovery();
      alert('Recovery acionado com sucesso!');
      await fetchData();
    } catch (error) {
      alert('Erro ao acionar recovery: ' + error);
    }
  };

  // Auto-refresh when monitoring
  useEffect(() => {
    if (!isMonitoring) return;

    const interval = setInterval(fetchData, 2000); // Atualiza a cada 2 segundos
    return () => clearInterval(interval);
  }, [isMonitoring]);

  if (!stats || !health) {
    return (
      <div className="dashboard-loading">
        <p>Carregando dashboard...</p>
      </div>
    );
  }

  return (
    <div className="orchestrator-dashboard">
      <div className="dashboard-header">
        <h2>📊 Dashboard do Orquestrador</h2>
        <div className="dashboard-controls">
          <button 
            className={`btn btn-primary ${isMonitoring ? 'active' : ''}`}
            onClick={toggleMonitoring}
          >
            {isMonitoring ? '⏸️ Pausar' : '▶️ Iniciar Monitoramento'}
          </button>
          <button className="btn btn-secondary" onClick={clearCache}>
            🗑️ Limpar Cache
          </button>
          <button className="btn btn-warning" onClick={triggerRecovery}>
            🔄 Recovery
          </button>
          {onClose && (
            <button className="btn btn-danger" onClick={onClose}>
              ✖️ Fechar
            </button>
          )}
        </div>
      </div>

      <div className="dashboard-grid">
        {/* Stats Cards */}
        <div className="stats-row">
          <Card className="stat-card">
            <div className="stat-header">
              <span className="stat-icon">📊</span>
              <span className="stat-label">Total de Tarefas</span>
            </div>
            <div className="stat-value">{stats.totalTasks}</div>
          </Card>

          <Card className="stat-card">
            <div className="stat-header">
              <span className="stat-icon">✅</span>
              <span className="stat-label">Sucesso</span>
            </div>
            <div className="stat-value success">{stats.successTasks}</div>
            <div className="stat-rate">{stats.successRate.toFixed(1)}%</div>
          </Card>

          <Card className="stat-card">
            <div className="stat-header">
              <span className="stat-icon">❌</span>
              <span className="stat-label">Falhas</span>
            </div>
            <div className="stat-value failed">{stats.failedTasks}</div>
            <div className="stat-rate">
              {((stats.failedTasks / stats.totalTasks) * 100).toFixed(1)}%
            </div>
          </Card>

          <Card className="stat-card">
            <div className="stat-header">
              <span className="stat-icon">⚙️</span>
              <span className="stat-label">Workers Ativos</span>
            </div>
            <div className="stat-value">{stats.activeWorkers}/5</div>
            <div className="stat-rate">
              {((stats.activeWorkers / 5) * 100).toFixed(0)}% uso
            </div>
          </Card>

          <Card className="stat-card">
            <div className="stat-header">
              <span className="stat-icon">⏱️</span>
              <span className="stat-label">Tempo Médio</span>
            </div>
            <div className="stat-value">{stats.avgTaskTime}</div>
            <div className="stat-rate">por tarefa</div>
          </Card>

          <Card className={`stat-card ${health.isHealthy ? 'healthy' : 'unhealthy'}`}>
            <div className="stat-header">
              <span className="stat-icon">{health.isHealthy ? '💚' : '❤️'}</span>
              <span className="stat-label">Status</span>
            </div>
            <div className="stat-value">
              {health.isHealthy ? 'Saudável' : 'Problemas'}
            </div>
            <div className="stat-rate">
              {health.lastCheck.split(' ')[1]}
            </div>
          </Card>
        </div>

        {/* Health Issues */}
        {!health.isHealthy && health.issues.length > 0 && (
          <Card className="health-issues-card">
            <h3>⚠️ Problemas Detectados</h3>
            <ul>
              {health.issues.map((issue, idx) => (
                <li key={idx}>{issue}</li>
              ))}
            </ul>
          </Card>
        )}

        {/* Charts */}
        <div className="charts-row">
          <Card className="chart-card">
            <h3>Taxa de Sucesso</h3>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={history}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" />
                <YAxis domain={[0, 100]} />
                <Tooltip />
                <Legend />
                <Line 
                  type="monotone" 
                  dataKey="successRate" 
                  stroke="#10b981" 
                  strokeWidth={2}
                  name="Taxa de Sucesso (%)"
                />
              </LineChart>
            </ResponsiveContainer>
          </Card>

          <Card className="chart-card">
            <h3>Workers Ativos</h3>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={history}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" />
                <YAxis domain={[0, 5]} />
                <Tooltip />
                <Legend />
                <Line 
                  type="monotone" 
                  dataKey="activeWorkers" 
                  stroke="#3b82f6" 
                  strokeWidth={2}
                  name="Workers Ativos"
                />
              </LineChart>
            </ResponsiveContainer>
          </Card>
        </div>

        <div className="charts-row">
          <Card className="chart-card">
            <h3>Tarefas Processadas</h3>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={history.slice(-20)}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar dataKey="successTasks" fill="#10b981" name="Sucesso" />
                <Bar dataKey="failedTasks" fill="#ef4444" name="Falhas" />
              </BarChart>
            </ResponsiveContainer>
          </Card>

          <Card className="chart-card">
            <h3>Tarefas Pendentes</h3>
            <div className="pending-info">
              <p className="pending-count">{health.tasksPending}</p>
              <p className="pending-label">tarefas na fila</p>
            </div>
            <div className="pending-bars">
              {Array.from({ length: 5 }, (_, i) => (
                <div 
                  key={i}
                  className={`pending-bar ${i < stats.activeWorkers ? 'active' : ''}`}
                />
              ))}
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
};