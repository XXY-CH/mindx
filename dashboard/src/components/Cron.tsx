import { useState, useEffect } from 'react';
import './styles/Cron.css';

interface Job {
  id: string;
  name: string;
  cron: string;
  message: string;
  command: string;
  enabled: boolean;
  created_at: string;
  last_run?: string;
  last_status: 'pending' | 'running' | 'success' | 'error';
  last_error?: string;
}

interface JobsResponse {
  jobs: Job[];
}

export default function Cron() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showDialog, setShowDialog] = useState(false);
  const [editingJob, setEditingJob] = useState<Job | null>(null);
  const [formData, setFormData] = useState<Partial<Job>>({
    name: '',
    cron: '',
    message: '',
    command: '',
    enabled: true,
  });
  const [actionLoading, setActionLoading] = useState(false);
  const [actionMessage, setActionMessage] = useState('');

  useEffect(() => {
    fetchJobs();
  }, []);

  const fetchJobs = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/cron/jobs');
      if (!response.ok) {
        throw new Error('Failed to fetch jobs');
      }
      const data: JobsResponse = await response.json();
      setJobs(data.jobs || []);
    } catch (error) {
      console.error('Failed to fetch jobs:', error);
      setError('加载定时任务失败');
    } finally {
      setLoading(false);
    }
  };

  const handleAdd = () => {
    setEditingJob(null);
    setFormData({
      name: '',
      cron: '',
      message: '',
      command: '',
      enabled: true,
    });
    setShowDialog(true);
  };

  const handleEdit = (job: Job) => {
    setEditingJob(job);
    setFormData({ ...job });
    setShowDialog(true);
  };

  const handleDelete = async (job: Job) => {
    if (!confirm(`确定要删除定时任务 "${job.name}" 吗？`)) {
      return;
    }

    try {
      setActionLoading(true);
      setActionMessage('删除中...');

      const response = await fetch(`/api/cron/jobs/${job.id}`, {
        method: 'DELETE',
      });
      if (!response.ok) {
        throw new Error('Failed to delete job');
      }

      alert('删除成功');
      fetchJobs();
    } catch (error) {
      console.error('Failed to delete job:', error);
      alert('删除失败');
    } finally {
      setActionLoading(false);
      setActionMessage('');
    }
  };

  const handleTogglePause = async (job: Job) => {
    const action = job.enabled ? 'pause' : 'resume';
    try {
      setActionLoading(true);
      setActionMessage(`${action === 'pause' ? '暂停' : '恢复'}中...`);

      const response = await fetch(`/api/cron/jobs/${job.id}/${action}`, {
        method: 'POST',
      });
      if (!response.ok) {
        throw new Error(`Failed to ${action} job`);
      }

      fetchJobs();
    } catch (error) {
      console.error(`Failed to ${action} job:`, error);
      alert(`${action === 'pause' ? '暂停' : '恢复'}失败`);
    } finally {
      setActionLoading(false);
      setActionMessage('');
    }
  };

  const handleSave = async () => {
    try {
      setActionLoading(true);
      setActionMessage('保存中...');

      const isEdit = editingJob !== null;
      const url = isEdit ? `/api/cron/jobs/${editingJob.id}` : '/api/cron/jobs';
      const method = isEdit ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });
      if (!response.ok) {
        throw new Error('Failed to save job');
      }

      alert('保存成功');
      setShowDialog(false);
      fetchJobs();
    } catch (error) {
      console.error('Failed to save job:', error);
      alert('保存失败');
    } finally {
      setActionLoading(false);
      setActionMessage('');
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success': return '✅';
      case 'running': return '🔄';
      case 'error': return '❌';
      default: return '⏳';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'success': return '成功';
      case 'running': return '运行中';
      case 'error': return '错误';
      default: return '等待中';
    }
  };

  if (loading) {
    return (
      <div className="cron-container">
        <div className="loading">加载中...</div>
      </div>
    );
  }

  return (
    <div className="cron-container">
      <div className="cron-header">
        <h1>任务</h1>
        <div className="header-actions">
          <button className="action-btn secondary" onClick={fetchJobs}>
            刷新
          </button>
          <button className="action-btn primary" onClick={handleAdd}>
            添加任务
          </button>
        </div>
      </div>

      {error && (
        <div className="error-message">{error}</div>
      )}

      <div className="cron-content">
        {jobs.length === 0 ? (
          <div className="empty-state">
            <p>暂无定时任务</p>
            <small>点击"添加任务"创建第一个定时任务</small>
          </div>
        ) : (
          <div className="jobs-list">
            {jobs.map((job) => (
              <div key={job.id} className="job-card">
                <div className="job-header">
                  <div className="job-title">
                    <h3>{job.name}</h3>
                    <span className={`job-enabled ${job.enabled ? 'enabled' : 'disabled'}`}>
                      {job.enabled ? '已启用' : '已暂停'}
                    </span>
                  </div>
                  <div className="job-status">
                    <span className="status-badge">
                      {getStatusIcon(job.last_status)} {getStatusText(job.last_status)}
                    </span>
                  </div>
                </div>

                <div className="job-details">
                  <div className="detail-item">
                    <label>Cron 表达式:</label>
                    <span>{job.cron}</span>
                  </div>
                  {job.message && (
                    <div className="detail-item">
                      <label>消息:</label>
                      <span>{job.message}</span>
                    </div>
                  )}
                  {job.command && (
                    <div className="detail-item">
                      <label>命令:</label>
                      <span>{job.command}</span>
                    </div>
                  )}
                  <div className="detail-item">
                    <label>创建时间:</label>
                    <span>{new Date(job.created_at).toLocaleString()}</span>
                  </div>
                  {job.last_run && (
                    <div className="detail-item">
                      <label>最后运行:</label>
                      <span>{new Date(job.last_run).toLocaleString()}</span>
                    </div>
                  )}
                  {job.last_error && (
                    <div className="detail-item error">
                      <label>错误:</label>
                      <span>{job.last_error}</span>
                    </div>
                  )}
                </div>

                <div className="job-actions">
                  <button
                    className="action-btn secondary"
                    onClick={() => handleEdit(job)}
                    disabled={actionLoading}
                  >
                    编辑
                  </button>
                  <button
                    className={`action-btn ${job.enabled ? 'warning' : 'success'}`}
                    onClick={() => handleTogglePause(job)}
                    disabled={actionLoading}
                  >
                    {job.enabled ? '暂停' : '恢复'}
                  </button>
                  <button
                    className="action-btn danger"
                    onClick={() => handleDelete(job)}
                    disabled={actionLoading}
                  >
                    删除
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {showDialog && (
        <div className="dialog-overlay" onClick={() => setShowDialog(false)}>
          <div className="dialog" onClick={(e) => e.stopPropagation()}>
            <h2>{editingJob ? '编辑定时任务' : '添加定时任务'}</h2>

            <div className="dialog-section">
              <div className="form-item">
                <label>任务名称 *</label>
                <input
                  type="text"
                  value={formData.name || ''}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="输入任务名称"
                />
              </div>

              <div className="form-item">
                <label>Cron 表达式 *</label>
                <input
                  type="text"
                  value={formData.cron || ''}
                  onChange={(e) => setFormData({ ...formData, cron: e.target.value })}
                  placeholder="例如: 0 9 * * 6"
                />
                <small>格式：分 时 日 月 周</small>
              </div>

              <div className="form-item">
                <label>消息 *</label>
                <input
                  type="text"
                  value={formData.message || ''}
                  onChange={(e) => setFormData({ ...formData, message: e.target.value })}
                  placeholder="例如: 帮我写日报"
                />
                <small>定时执行时会发送这条消息</small>
              </div>

              <div className="form-item">
                <label>命令（可选）</label>
                <input
                  type="text"
                  value={formData.command || ''}
                  onChange={(e) => setFormData({ ...formData, command: e.target.value })}
                  placeholder="输入要执行的命令（可选）"
                />
              </div>

              <div className="form-item">
                <label>
                  <input
                    type="checkbox"
                    checked={formData.enabled || false}
                    onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                  />
                  启用任务
                </label>
              </div>
            </div>

            <div className="dialog-actions">
              <button
                className="action-btn secondary"
                onClick={() => setShowDialog(false)}
                disabled={actionLoading}
              >
                取消
              </button>
              <button
                className="action-btn primary"
                onClick={handleSave}
                disabled={actionLoading}
              >
                {actionLoading ? '保存中...' : '保存'}
              </button>
            </div>

            {actionMessage && <div className="action-message">{actionMessage}</div>}
          </div>
        </div>
      )}

      {actionLoading && (
        <div className="loading-overlay">
          <div className="loading-spinner"></div>
          <p>{actionMessage || '处理中...'}</p>
        </div>
      )}
    </div>
  );
}
