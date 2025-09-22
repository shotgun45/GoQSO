import React, { useState } from 'react';
import { Download, Calendar } from 'lucide-react';
import { qsoApi, downloadFile } from '../api';

interface ExportFormProps {
  onError: (error: string) => void;
}

const ExportForm: React.FC<ExportFormProps> = ({ onError }) => {
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');
  const [isExporting, setIsExporting] = useState(false);
  const [exportOption, setExportOption] = useState<'all' | 'range'>('all');

  const handleExport = async () => {
    if (exportOption === 'range' && (!startDate || !endDate)) {
      onError('Please specify both start and end dates for date range export');
      return;
    }

    if (exportOption === 'range' && startDate > endDate) {
      onError('Start date must be before or equal to end date');
      return;
    }

    setIsExporting(true);
    try {
      const blob = await qsoApi.exportADIF(
        exportOption === 'range' ? startDate : undefined,
        exportOption === 'range' ? endDate : undefined
      );
      
      // Generate filename based on export type
      let filename = 'goqso_export';
      if (exportOption === 'range') {
        filename += `_${startDate}_to_${endDate}`;
      } else {
        filename += `_${new Date().toISOString().split('T')[0]}`;
      }
      filename += '.adi';
      
      downloadFile(blob, filename);
    } catch (err) {
      onError('Failed to export ADIF: ' + (err instanceof Error ? err.message : 'Unknown error'));
    } finally {
      setIsExporting(false);
    }
  };

  const resetDates = () => {
    setStartDate('');
    setEndDate('');
  };

  return (
    <div className="export-form">
      <div className="form-header">
        <Download size={24} />
        <h2>Export QSO Log</h2>
      </div>
      
      <div className="form-content">
        <div className="export-options">
          <div className="radio-group">
            <label className={`radio-option ${exportOption === 'all' ? 'selected' : ''}`}>
              <input
                type="radio"
                name="exportOption"
                value="all"
                checked={exportOption === 'all'}
                onChange={(e) => setExportOption(e.target.value as 'all' | 'range')}
              />
              <span className="radio-label">
                <strong>Export All Contacts</strong>
                <br />
                <small>Export your entire QSO log</small>
              </span>
            </label>
            
            <label className={`radio-option ${exportOption === 'range' ? 'selected' : ''}`}>
              <input
                type="radio"
                name="exportOption"
                value="range"
                checked={exportOption === 'range'}
                onChange={(e) => setExportOption(e.target.value as 'all' | 'range')}
              />
              <span className="radio-label">
                <strong>Export Date Range</strong>
                <br />
                <small>Export contacts within a specific date range</small>
              </span>
            </label>
          </div>
        </div>

        {exportOption === 'range' && (
          <div className="date-range">
            <div className="date-inputs">
              <div className="input-group">
                <label htmlFor="startDate">
                  <Calendar size={16} />
                  Start Date
                </label>
                <input
                  type="date"
                  id="startDate"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  max={endDate || undefined}
                />
              </div>
              
              <div className="input-group">
                <label htmlFor="endDate">
                  <Calendar size={16} />
                  End Date
                </label>
                <input
                  type="date"
                  id="endDate"
                  value={endDate}
                  onChange={(e) => setEndDate(e.target.value)}
                  min={startDate || undefined}
                />
              </div>
            </div>
            
            <button 
              type="button" 
              onClick={resetDates}
              className="reset-dates-btn"
            >
              Clear Dates
            </button>
          </div>
        )}

        <div className="export-info">
          <div className="info-box">
            <h3>About ADIF Export</h3>
            <p>
              ADIF (Amateur Data Interchange Format) is the standard format for amateur radio logbook data exchange. 
              The exported file can be imported into other logging software or used for award applications.
            </p>
            <ul>
              <li>Compatible with most amateur radio logging software</li>
              <li>Includes all contact details (callsign, date, time, band, mode, etc.)</li>
              <li>Preserves QSL confirmation status</li>
              <li>Ready for LOTW, eQSL, or contest submissions</li>
            </ul>
          </div>
        </div>

        <div className="form-actions">
          <button 
            type="button"
            onClick={handleExport}
            disabled={isExporting}
            className="export-btn primary"
          >
            <Download size={16} />
            {isExporting ? 'Exporting...' : 'Export ADIF File'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ExportForm;