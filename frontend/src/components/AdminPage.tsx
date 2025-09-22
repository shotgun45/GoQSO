import React, { useState, useEffect } from 'react';
import { Settings, Database, RefreshCw, Info, AlertTriangle, Download, Merge } from 'lucide-react';

interface SystemInfo {
  application: {
    name: string;
    version: string;
    go_version: string;
    start_time: string;
    uptime: string;
  };
  database: {
    contact_count: number;
    database_size: number;
    duplicate_count: number;
    open_connections: number;
    max_open_conns: number;
    idle_connections: number;
    in_use: number;
    wait_count: number;
    wait_duration: string;
    max_idle_closed: number;
    max_idle_time_closed: number;
    max_lifetime_closed: number;
  };
}

interface AdminPageProps {
  onError: (error: string) => void;
}

const AdminPage: React.FC<AdminPageProps> = ({ onError }) => {
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [lastRefresh, setLastRefresh] = useState<Date | null>(null);
  const [showMergeModal, setShowMergeModal] = useState(false);
  const [mergeStep, setMergeStep] = useState<'warning' | 'export' | 'confirm' | 'processing' | 'complete'>('warning');
  const [mergeResult, setMergeResult] = useState<{ merged_count: number; message: string } | null>(null);
  const [mergeLoading, setMergeLoading] = useState(false);
  const [confirmInput, setConfirmInput] = useState('');

  const fetchSystemInfo = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/admin/system');
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const data = await response.json();
      if (data.success) {
        setSystemInfo(data.data);
        setLastRefresh(new Date());
      } else {
        throw new Error(data.error || 'Failed to fetch system information');
      }
    } catch (err) {
      onError('Failed to load system information: ' + (err instanceof Error ? err.message : 'Unknown error'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSystemInfo();
  }, []);

  const formatUptime = (uptime: string): string => {
    // Parse Go duration string and make it more readable
    const match = uptime.match(/^(?:(\d+)h)?(?:(\d+)m)?(?:(\d+(?:\.\d+)?)s)?$/);
    if (!match) return uptime;
    
    const hours = parseInt(match[1] || '0');
    const minutes = parseInt(match[2] || '0');
    const seconds = parseFloat(match[3] || '0');
    
    const parts = [];
    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);
    if (seconds > 0 && hours === 0) parts.push(`${Math.floor(seconds)}s`);
    
    return parts.join(' ') || '< 1s';
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  const handleExportBeforeMerge = async () => {
    try {
      const response = await fetch('/api/contacts/export');
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.style.display = 'none';
      a.href = url;
      a.download = `goqso-backup-${new Date().toISOString().split('T')[0]}.adi`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
      setMergeStep('confirm');
    } catch (err) {
      onError('Failed to export backup: ' + (err instanceof Error ? err.message : 'Unknown error'));
    }
  };

  const handleMergeDuplicates = async () => {
    setMergeLoading(true);
    setMergeStep('processing');
    
    try {
      const response = await fetch('/api/admin/merge-duplicates', {
        method: 'POST',
      });
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const data = await response.json();
      if (data.success) {
        setMergeResult(data.data);
        setMergeStep('complete');
        // Refresh system info to update contact count
        fetchSystemInfo();
      } else {
        throw new Error(data.error || 'Failed to merge duplicates');
      }
    } catch (err) {
      onError('Failed to merge duplicates: ' + (err instanceof Error ? err.message : 'Unknown error'));
      setShowMergeModal(false);
      setMergeStep('warning');
    } finally {
      setMergeLoading(false);
    }
  };

  const resetMergeModal = () => {
    setShowMergeModal(false);
    setMergeStep('warning');
    setMergeResult(null);
    setMergeLoading(false);
    setConfirmInput('');
  };

  const StatCard: React.FC<{
    title: string;
    icon: React.ComponentType<any>;
    children: React.ReactNode;
    className?: string;
  }> = ({ title, icon: Icon, children, className = '' }) => (
    <div className={`admin-stat-card ${className}`}>
      <div className="stat-card-header">
        <Icon size={20} />
        <h3>{title}</h3>
      </div>
      <div className="stat-card-content">
        {children}
      </div>
    </div>
  );

  const StatRow: React.FC<{ label: string; value: string | number; unit?: string }> = ({ 
    label, 
    value, 
    unit = '' 
  }) => (
    <div className="stat-row">
      <span className="stat-label">{label}</span>
      <span className="stat-value">
        {typeof value === 'number' ? value.toLocaleString() : value}
        {unit && <span className="stat-unit">{unit}</span>}
      </span>
    </div>
  );

  return (
    <div className="admin-page">
      <div className="admin-header">
        <div className="admin-title">
          <Settings size={24} />
          <h2>System Administration</h2>
        </div>
        <div className="admin-actions">
          {lastRefresh && (
            <span className="last-refresh">
              Last updated: {lastRefresh.toLocaleTimeString()}
            </span>
          )}
          <button 
            onClick={fetchSystemInfo} 
            disabled={loading}
            className="refresh-btn"
          >
            <RefreshCw size={16} className={loading ? 'spinning' : ''} />
            Refresh
          </button>
        </div>
      </div>

      {/* Duplicate Records Notification */}
      {systemInfo && systemInfo.database.duplicate_count > 0 && (
        <div className="admin-notification">
          <div className="notification-box warning">
            <AlertTriangle size={20} className="notification-icon" />
            <div className="notification-content">
              <h4>Duplicate Records Detected</h4>
              <p>
                Found <strong>{systemInfo.database.duplicate_count}</strong> duplicate contact{systemInfo.database.duplicate_count !== 1 ? 's' : ''} in your database.
                Use the merge tool below to clean up your data.
              </p>
            </div>
          </div>
        </div>
      )}

      {systemInfo && (
        <div className="admin-content">
          <div className="admin-grid">
            <StatCard title="Application Info" icon={Info}>
              <StatRow label="Name" value={systemInfo.application.name} />
              <StatRow label="Version" value={systemInfo.application.version} />
              <StatRow label="Go Version" value={systemInfo.application.go_version} />
              <StatRow label="Uptime" value={formatUptime(systemInfo.application.uptime)} />
              <StatRow 
                label="Started" 
                value={new Date(systemInfo.application.start_time).toLocaleString()} 
              />
            </StatCard>

            <StatCard title="Database Status" icon={Database}>
              <StatRow 
                label="Total Contacts" 
                value={systemInfo.database.contact_count} 
                unit="QSOs" 
              />
              <StatRow 
                label="Database Size" 
                value={formatBytes(systemInfo.database.database_size)} 
              />
            </StatCard>
          </div>

          <div className="admin-info-section">
            <div className="info-box">
              <h3>Database Management</h3>
              <p>
                Manage your contact database with tools for cleaning and optimization.
              </p>
              <div className="admin-actions-grid">
                <button 
                  onClick={() => setShowMergeModal(true)}
                  className="admin-action-btn merge-btn"
                  title="Merge duplicate contact records"
                >
                  <Merge size={20} />
                  Merge Duplicates
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Merge Duplicates Modal */}
      {showMergeModal && (
        <div className="modal-overlay">
          <div className="modal-content merge-modal">
            {mergeStep === 'warning' && (
              <>
                <div className="modal-header">
                  <AlertTriangle size={24} className="warning-icon" />
                  <h3>Merge Duplicate Records</h3>
                </div>
                <div className="modal-body">
                  <div className="warning-box">
                    <p><strong>⚠️ Data Loss Warning</strong></p>
                    <p>
                      This operation will merge duplicate contact records based on callsign, date, and time. 
                      The oldest record will be kept and data from duplicates will be merged where possible.
                    </p>
                    <ul>
                      <li>Duplicate records will be permanently deleted</li>
                      <li>Some data may be lost if conflicts cannot be resolved</li>
                      <li>This operation cannot be undone</li>
                    </ul>
                  </div>
                  <p><strong>We strongly recommend exporting your data before proceeding.</strong></p>
                </div>
                <div className="modal-actions">
                  <button onClick={resetMergeModal} className="btn-secondary">
                    Cancel
                  </button>
                  <button onClick={() => setMergeStep('export')} className="btn-warning">
                    Continue
                  </button>
                </div>
              </>
            )}

            {mergeStep === 'export' && (
              <>
                <div className="modal-header">
                  <Download size={24} />
                  <h3>Export Backup</h3>
                </div>
                <div className="modal-body">
                  <p>Before merging duplicates, we recommend creating a backup of your current data.</p>
                  <p>Click "Export Backup" to download your complete contact list as an ADIF file.</p>
                </div>
                <div className="modal-actions">
                  <button onClick={resetMergeModal} className="btn-secondary">
                    Cancel
                  </button>
                  <button onClick={() => setMergeStep('confirm')} className="btn-secondary">
                    Skip Backup
                  </button>
                  <button onClick={handleExportBeforeMerge} className="btn-primary">
                    Export Backup
                  </button>
                </div>
              </>
            )}

            {mergeStep === 'confirm' && (
              <>
                <div className="modal-header">
                  <AlertTriangle size={24} className="warning-icon" />
                  <h3>Final Confirmation</h3>
                </div>
                <div className="modal-body">
                  <p><strong>Are you sure you want to merge duplicate records?</strong></p>
                  <p>This action will:</p>
                  <ul>
                    <li>Identify contacts with the same callsign, date, and time</li>
                    <li>Keep the oldest record and merge additional data</li>
                    <li>Permanently delete duplicate records</li>
                  </ul>
                  <p className="confirm-text">
                    Type "MERGE" to confirm you understand this action cannot be undone:
                  </p>
                  <input 
                    type="text" 
                    placeholder="Type MERGE to confirm"
                    value={confirmInput}
                    onChange={(e) => setConfirmInput(e.target.value)}
                    className="confirm-input"
                    disabled={mergeLoading}
                  />
                </div>
                <div className="modal-actions">
                  <button onClick={resetMergeModal} className="btn-secondary" disabled={mergeLoading}>
                    Cancel
                  </button>
                  <button 
                    onClick={handleMergeDuplicates} 
                    className="btn-danger confirm-merge-btn"
                    disabled={mergeLoading || confirmInput !== 'MERGE'}
                  >
                    {mergeLoading ? 'Processing...' : 'Merge Duplicates'}
                  </button>
                </div>
              </>
            )}

            {mergeStep === 'processing' && (
              <>
                <div className="modal-header">
                  <RefreshCw size={24} className="spinning" />
                  <h3>Processing...</h3>
                </div>
                <div className="modal-body">
                  <p>Merging duplicate records. Please wait...</p>
                </div>
              </>
            )}

            {mergeStep === 'complete' && mergeResult && (
              <>
                <div className="modal-header">
                  <Database size={24} />
                  <h3>Merge Complete</h3>
                </div>
                <div className="modal-body">
                  <div className="success-box">
                    <p><strong>✅ Merge Successful</strong></p>
                    <p>{mergeResult.message}</p>
                    {mergeResult.merged_count === 0 && (
                      <p><em>No duplicate records were found.</em></p>
                    )}
                  </div>
                </div>
                <div className="modal-actions">
                  <button onClick={resetMergeModal} className="btn-primary">
                    Close
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {loading && !systemInfo && (
        <div className="admin-loading">
          <RefreshCw size={32} className="spinning" />
          <p>Loading system information...</p>
        </div>
      )}
    </div>
  );
};

export default AdminPage;