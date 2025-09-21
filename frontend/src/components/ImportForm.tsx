import React, { useState } from 'react';
import { ImportResult, ImportOptions, LotwCredentials } from '../types';
import { 
  Upload, 
  FileText, 
  Globe, 
  CheckCircle, 
  Info,
  Settings,
  Calendar
} from 'lucide-react';

interface ImportFormProps {
  onImportComplete: (result: ImportResult) => void;
}

const ImportForm: React.FC<ImportFormProps> = ({ onImportComplete }) => {
  const [importType, setImportType] = useState<'adif' | 'lotw'>('adif');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);
  const [importOptions, setImportOptions] = useState<ImportOptions>({
    file_type: 'adif',
    merge_duplicates: true,
    update_existing: false,
  });
  const [lotwCredentials, setLotwCredentials] = useState<LotwCredentials>({
    username: '',
    password: '',
    start_date: '',
    end_date: '',
  });
  const [dragOver, setDragOver] = useState(false);

  const handleFileSelect = (file: File) => {
    const validTypes = ['.adi', '.adif', '.txt'];
    const fileExt = file.name.toLowerCase().substring(file.name.lastIndexOf('.'));
    
    if (!validTypes.includes(fileExt)) {
      alert('Please select a valid ADIF file (.adi, .adif, or .txt)');
      return;
    }
    
    setSelectedFile(file);
  };

  const handleFileInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const handleDragOver = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setDragOver(true);
  };

  const handleDragLeave = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setDragOver(false);
  };

  const handleDrop = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setDragOver(false);
    
    const file = event.dataTransfer.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const handleAdifImport = async () => {
    if (!selectedFile) return;

    setImporting(true);
    try {
      const formData = new FormData();
      formData.append('file', selectedFile);
      formData.append('options', JSON.stringify(importOptions));

      const response = await fetch('/api/import/adif', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result: ImportResult = await response.json();
      onImportComplete(result);
    } catch (error) {
      onImportComplete({
        success: false,
        imported_count: 0,
        skipped_count: 0,
        error_count: 1,
        errors: [error instanceof Error ? error.message : 'Unknown error occurred'],
        message: 'Import failed',
      });
    } finally {
      setImporting(false);
    }
  };

  const handleLotwImport = async () => {
    if (!lotwCredentials.username || !lotwCredentials.password) {
      alert('Please enter your LoTW username and password');
      return;
    }

    setImporting(true);
    try {
      const response = await fetch('/api/import/lotw', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          credentials: lotwCredentials,
          options: { ...importOptions, file_type: 'lotw' },
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result: ImportResult = await response.json();
      onImportComplete(result);
    } catch (error) {
      onImportComplete({
        success: false,
        imported_count: 0,
        skipped_count: 0,
        error_count: 1,
        errors: [error instanceof Error ? error.message : 'Unknown error occurred'],
        message: 'LoTW import failed',
      });
    } finally {
      setImporting(false);
    }
  };

  const handleImport = () => {
    if (importType === 'adif') {
      handleAdifImport();
    } else {
      handleLotwImport();
    }
  };

  return (
    <div className="import-form">
      <div className="import-header">
        <h2>Import QSO Data</h2>
        <p>Import contacts from ADIF files or Logbook of the World</p>
      </div>

      <div className="import-type-selector">
        <div className="radio-group">
          <label className="radio-option">
            <input
              type="radio"
              name="importType"
              value="adif"
              checked={importType === 'adif'}
              onChange={(e) => setImportType(e.target.value as 'adif' | 'lotw')}
            />
            <div className="radio-content">
              <div className="radio-icon">
                <FileText size={24} />
              </div>
              <div className="radio-text">
                <h3>ADIF File</h3>
                <p>Import from Amateur Data Interchange Format files (.adi, .adif)</p>
              </div>
            </div>
          </label>

          <label className="radio-option">
            <input
              type="radio"
              name="importType"
              value="lotw"
              checked={importType === 'lotw'}
              onChange={(e) => setImportType(e.target.value as 'adif' | 'lotw')}
            />
            <div className="radio-content">
              <div className="radio-icon">
                <Globe size={24} />
              </div>
              <div className="radio-text">
                <h3>Logbook of the World</h3>
                <p>Import confirmed QSOs directly from ARRL LoTW</p>
              </div>
            </div>
          </label>
        </div>
      </div>

      {importType === 'adif' && (
        <div className="adif-import-section">
          <div className="file-upload-area">
            <div
              className={`file-drop-zone ${dragOver ? 'drag-over' : ''} ${selectedFile ? 'has-file' : ''}`}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => document.getElementById('file-input')?.click()}
            >
              <input
                id="file-input"
                type="file"
                accept=".adi,.adif,.txt"
                onChange={handleFileInputChange}
                style={{ display: 'none' }}
              />
              
              {selectedFile ? (
                <div className="file-selected">
                  <CheckCircle size={48} className="file-icon success" />
                  <h3>{selectedFile.name}</h3>
                  <p>{(selectedFile.size / 1024).toFixed(1)} KB</p>
                  <button 
                    className="change-file-btn"
                    onClick={(e) => {
                      e.stopPropagation();
                      setSelectedFile(null);
                    }}
                  >
                    Change File
                  </button>
                </div>
              ) : (
                <div className="file-prompt">
                  <Upload size={48} className="file-icon" />
                  <h3>Drop ADIF file here or click to browse</h3>
                  <p>Supports .adi, .adif, and .txt files</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {importType === 'lotw' && (
        <div className="lotw-import-section">
          <div className="lotw-credentials">
            <h3>
              <Globe size={20} />
              LoTW Credentials
            </h3>
            <div className="form-row">
              <div className="form-group">
                <label htmlFor="lotw-username">Username</label>
                <input
                  id="lotw-username"
                  type="text"
                  value={lotwCredentials.username}
                  onChange={(e) => setLotwCredentials(prev => ({ ...prev, username: e.target.value }))}
                  placeholder="Your LoTW username"
                />
              </div>
              <div className="form-group">
                <label htmlFor="lotw-password">Password</label>
                <input
                  id="lotw-password"
                  type="password"
                  value={lotwCredentials.password}
                  onChange={(e) => setLotwCredentials(prev => ({ ...prev, password: e.target.value }))}
                  placeholder="Your LoTW password"
                />
              </div>
            </div>
            
            <div className="security-notice">
              <div className="warning-icon">⚠️</div>
              <div className="warning-text">
                <strong>Security Notice:</strong> LoTW API requires credentials to be sent as URL parameters, 
                which means they will appear in your browser's network tab during the import process. 
                This is a limitation of the LoTW API design.
              </div>
            </div>
            
            <div className="form-row">
              <div className="form-group">
                <label htmlFor="lotw-start-date">
                  <Calendar size={16} />
                  Start Date (optional)
                </label>
                <input
                  id="lotw-start-date"
                  type="date"
                  value={lotwCredentials.start_date}
                  onChange={(e) => setLotwCredentials(prev => ({ ...prev, start_date: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label htmlFor="lotw-end-date">
                  <Calendar size={16} />
                  End Date (optional)
                </label>
                <input
                  id="lotw-end-date"
                  type="date"
                  value={lotwCredentials.end_date}
                  onChange={(e) => setLotwCredentials(prev => ({ ...prev, end_date: e.target.value }))}
                />
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="import-options">
        <h3>
          <Settings size={20} />
          Import Options
        </h3>
        
        <div className="options-grid">
          <label className="checkbox-option">
            <input
              type="checkbox"
              checked={importOptions.merge_duplicates}
              onChange={(e) => setImportOptions(prev => ({ ...prev, merge_duplicates: e.target.checked }))}
            />
            <div className="checkbox-content">
              <h4>Merge Duplicates</h4>
              <p>Skip contacts that already exist (based on callsign and date/time)</p>
            </div>
          </label>

          <label className="checkbox-option">
            <input
              type="checkbox"
              checked={importOptions.update_existing}
              onChange={(e) => setImportOptions(prev => ({ ...prev, update_existing: e.target.checked }))}
            />
            <div className="checkbox-content">
              <h4>Update Existing</h4>
              <p>Update existing contacts with new information from import</p>
            </div>
          </label>
        </div>

        <div className="import-info">
          <Info size={16} />
          <p>
            {importType === 'adif' 
              ? 'ADIF files are the standard format for amateur radio logging. Most logging software can export to this format.'
              : 'LoTW import will only retrieve confirmed QSOs from your Logbook of the World account.'
            }
          </p>
        </div>
      </div>

      <div className="import-actions">
        <button
          className="import-btn"
          onClick={handleImport}
          disabled={
            importing || 
            (importType === 'adif' && !selectedFile) ||
            (importType === 'lotw' && (!lotwCredentials.username || !lotwCredentials.password))
          }
        >
          {importing ? (
            <>
              <div className="spinner" />
              Importing...
            </>
          ) : (
            <>
              <Upload size={16} />
              Import {importType === 'adif' ? 'ADIF File' : 'from LoTW'}
            </>
          )}
        </button>
      </div>
    </div>
  );
};

export default ImportForm;