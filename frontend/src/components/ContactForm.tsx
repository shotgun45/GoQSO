import React, { useState, useCallback, FormEvent, useEffect } from 'react';
import { NewContact, Contact } from '../types';
import { Save, X } from 'lucide-react';

interface ContactFormProps {
  onSave: (contact: NewContact) => void;
  onCancel: () => void;
  loading?: boolean;
  editingContact?: Contact;
}

const ContactForm: React.FC<ContactFormProps> = ({ onSave, onCancel, loading = false, editingContact }) => {
  // Helper functions to convert between different time formats
  const formatTimeForInput = (timeString: string): string => {
    if (!timeString) return '';
    
    // If it's already in HH:MM or HH:MM:SS format, return HH:MM part for input
    if (timeString.includes(':')) {
      return timeString.slice(0, 5); // Take only HH:MM part for HTML time input
    }
    
    // If it's in HHMM format (like "1746"), convert to HH:MM
    if (timeString.length >= 4) {
      const hours = timeString.slice(0, 2);
      const minutes = timeString.slice(2, 4);
      return `${hours}:${minutes}`;
    }
    
    return timeString;
  };

  const formatTimeForDatabase = (timeString: string): string => {
    if (!timeString) return '';
    
    // If input is HH:MM format, add :00 seconds and keep as HH:MM:SS
    if (timeString.includes(':')) {
      // If it's HH:MM, add seconds
      if (timeString.length === 5) {
        return `${timeString}:00`;
      }
      // If it's already HH:MM:SS, return as is
      return timeString;
    }
    
    // If it's HHMM format, convert to HH:MM:SS
    if (timeString.length >= 4) {
      const hours = timeString.slice(0, 2);
      const minutes = timeString.slice(2, 4);
      return `${hours}:${minutes}:00`;
    }
    
    return timeString;
  };

  const getInitialFormData = (): NewContact => {
    if (editingContact) {
      return {
        callsign: editingContact.callsign,
        operator_name: editingContact.operator_name,
        contact_date: editingContact.contact_date.split('T')[0], // Convert ISO date to YYYY-MM-DD
        time_on: formatTimeForInput(editingContact.time_on),
        time_off: formatTimeForInput(editingContact.time_off),
        frequency: editingContact.frequency,
        band: editingContact.band,
        mode: editingContact.mode,
        power_watts: editingContact.power_watts,
        rst_sent: editingContact.rst_sent,
        rst_received: editingContact.rst_received,
        qth: editingContact.qth,
        country: editingContact.country,
        grid_square: editingContact.grid_square,
        comment: editingContact.comment,
        confirmed: editingContact.confirmed,
      };
    }
    
    return {
      callsign: '',
      operator_name: '',
      contact_date: new Date().toISOString().split('T')[0], // Today's date
      time_on: '',
      time_off: '',
      frequency: 0,
      band: '',
      mode: '',
      power_watts: 0,
      rst_sent: '59',
      rst_received: '59',
      qth: '',
      country: '',
      grid_square: '',
      comment: '',
      confirmed: false,
    };
  };

  const [formData, setFormData] = useState<NewContact>(getInitialFormData);

  // Reset form when editingContact changes
  useEffect(() => {
    setFormData(getInitialFormData());
    setErrors({});
  }, [editingContact]);

  type ContactFormErrors = {
    [K in keyof NewContact]?: string;
  };
  const [errors, setErrors] = useState<ContactFormErrors>({});

  // Common bands and modes for radio operators
  const bands = ['160m', '80m', '40m', '30m', '20m', '17m', '15m', '12m', '10m', '6m', '4m', '2m', '70cm'];
  const modes = ['SSB', 'CW', 'FT8', 'FT4', 'PSK31', 'RTTY', 'AM', 'FM', 'DIGITAL'];

  const validateForm = (): boolean => {
    const newErrors: ContactFormErrors = {};

    if (!formData.callsign.trim()) {
      newErrors.callsign = 'Callsign is required';
    } else if (!/^[A-Z0-9/]+$/i.test(formData.callsign)) {
      newErrors.callsign = 'Invalid callsign format';
    }

    if (!formData.contact_date) {
      newErrors.contact_date = 'Date is required';
    }

    if (!formData.time_on) {
      newErrors.time_on = 'Start time is required';
    }

    if (!formData.time_off) {
      newErrors.time_off = 'End time is required';
    }

    if (formData.time_on && formData.time_off && formData.time_on >= formData.time_off) {
      newErrors.time_off = 'End time must be after start time';
    }

    if (!formData.frequency || formData.frequency <= 0) {
      newErrors.frequency = 'Valid frequency is required';
    }

    if (!formData.band.trim()) {
      newErrors.band = 'Band is required';
    }

    if (!formData.mode.trim()) {
      newErrors.mode = 'Mode is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = useCallback((e: FormEvent) => {
    e.preventDefault();
    if (validateForm()) {
      onSave({
        ...formData,
        callsign: formData.callsign.toUpperCase(),
        operator_name: formData.operator_name.trim(),
        time_on: formatTimeForDatabase(formData.time_on),
        time_off: formatTimeForDatabase(formData.time_off),
        band: formData.band.trim(),
        mode: formData.mode.toUpperCase(),
        qth: formData.qth.trim(),
        country: formData.country.trim(),
        grid_square: formData.grid_square.toUpperCase(),
        comment: formData.comment.trim(),
      });
    }
  }, [formData, onSave]);

  const handleInputChange = (field: keyof NewContact, value: string | number | boolean) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const getCurrentTime = () => {
    const now = new Date();
    return now.toISOString().slice(11, 16); // Get HH:MM in UTC format
  };

  return (
    <div className="contact-form">
      <div className="form-header">
        <h2>{editingContact ? 'Edit QSO' : 'Log New QSO'}</h2>
        <button onClick={onCancel} className="close-btn" title="Cancel">
          <X size={20} />
        </button>
      </div>

      <form onSubmit={handleSubmit} className="form-content">
        <div className="form-row">
          <div className="form-group">
            <label htmlFor="callsign">Callsign *</label>
            <input
              id="callsign"
              type="text"
              value={formData.callsign}
              onChange={(e) => handleInputChange('callsign', e.target.value)}
              className={errors.callsign ? 'error' : ''}
              placeholder="W1ABC"
              disabled={loading}
              autoFocus
            />
            {errors.callsign && <span className="error-text">{errors.callsign}</span>}
          </div>

          <div className="form-group">
            <label htmlFor="operator_name">Operator Name</label>
            <input
              id="operator_name"
              type="text"
              value={formData.operator_name}
              onChange={(e) => handleInputChange('operator_name', e.target.value)}
              placeholder="John Doe"
              disabled={loading}
            />
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="contact_date">Date *</label>
            <input
              id="contact_date"
              type="date"
              value={formData.contact_date}
              onChange={(e) => handleInputChange('contact_date', e.target.value)}
              className={errors.contact_date ? 'error' : ''}
              disabled={loading}
            />
            {errors.contact_date && <span className="error-text">{errors.contact_date}</span>}
          </div>

          <div className="form-group">
            <label htmlFor="time_on">
              Start Time (UTC) *
              <button
                type="button"
                onClick={() => handleInputChange('time_on', getCurrentTime())}
                className="now-btn"
                disabled={loading}
              >
                Now
              </button>
            </label>
            <input
              id="time_on"
              type="time"
              step="1"
              value={formData.time_on}
              onChange={(e) => handleInputChange('time_on', e.target.value)}
              className={errors.time_on ? 'error' : ''}
              disabled={loading}
            />
            {errors.time_on && <span className="error-text">{errors.time_on}</span>}
          </div>

          <div className="form-group">
            <label htmlFor="time_off">
              End Time (UTC) *
              <button
                type="button"
                onClick={() => handleInputChange('time_off', getCurrentTime())}
                className="now-btn"
                disabled={loading}
              >
                Now
              </button>
            </label>
            <input
              id="time_off"
              type="time"
              step="1"
              value={formData.time_off}
              onChange={(e) => handleInputChange('time_off', e.target.value)}
              className={errors.time_off ? 'error' : ''}
              disabled={loading}
            />
            {errors.time_off && <span className="error-text">{errors.time_off}</span>}
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="frequency">Frequency (MHz) *</label>
            <input
              id="frequency"
              type="number"
              step="0.001"
              min="0.001"
              max="999.999"
              value={formData.frequency || ''}
              onChange={(e) => handleInputChange('frequency', parseFloat(e.target.value) || 0)}
              className={errors.frequency ? 'error' : ''}
              placeholder="14.205"
              disabled={loading}
            />
            {errors.frequency && <span className="error-text">{errors.frequency}</span>}
          </div>

          <div className="form-group">
            <label htmlFor="band">Band *</label>
            <select
              id="band"
              value={formData.band}
              onChange={(e) => handleInputChange('band', e.target.value)}
              className={errors.band ? 'error' : ''}
              disabled={loading}
            >
              <option value="">Select Band</option>
              {bands.map(band => (
                <option key={band} value={band}>{band}</option>
              ))}
            </select>
            {errors.band && <span className="error-text">{errors.band}</span>}
          </div>

          <div className="form-group">
            <label htmlFor="mode">Mode *</label>
            <select
              id="mode"
              value={formData.mode}
              onChange={(e) => handleInputChange('mode', e.target.value)}
              className={errors.mode ? 'error' : ''}
              disabled={loading}
            >
              <option value="">Select Mode</option>
              {modes.map(mode => (
                <option key={mode} value={mode}>{mode}</option>
              ))}
            </select>
            {errors.mode && <span className="error-text">{errors.mode}</span>}
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="rst_sent">RST Sent</label>
            <input
              id="rst_sent"
              type="text"
              value={formData.rst_sent}
              onChange={(e) => handleInputChange('rst_sent', e.target.value)}
              placeholder="59"
              maxLength={3}
              disabled={loading}
            />
          </div>

          <div className="form-group">
            <label htmlFor="rst_received">RST Received</label>
            <input
              id="rst_received"
              type="text"
              value={formData.rst_received}
              onChange={(e) => handleInputChange('rst_received', e.target.value)}
              placeholder="59"
              maxLength={3}
              disabled={loading}
            />
          </div>

          <div className="form-group">
            <label htmlFor="power_watts">Power (Watts)</label>
            <input
              id="power_watts"
              type="number"
              min="1"
              max="1500"
              value={formData.power_watts || ''}
              onChange={(e) => handleInputChange('power_watts', parseInt(e.target.value) || 0)}
              placeholder="Enter power in watts"
              disabled={loading}
            />
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="qth">QTH</label>
            <input
              id="qth"
              type="text"
              value={formData.qth}
              onChange={(e) => handleInputChange('qth', e.target.value)}
              placeholder="New York, NY"
              disabled={loading}
            />
          </div>

          <div className="form-group">
            <label htmlFor="country">Country</label>
            <input
              id="country"
              type="text"
              value={formData.country}
              onChange={(e) => handleInputChange('country', e.target.value)}
              placeholder="United States"
              disabled={loading}
            />
          </div>

          <div className="form-group">
            <label htmlFor="grid_square">Grid Square</label>
            <input
              id="grid_square"
              type="text"
              value={formData.grid_square}
              onChange={(e) => handleInputChange('grid_square', e.target.value)}
              placeholder="FN20"
              maxLength={6}
              disabled={loading}
            />
          </div>
        </div>

        <div className="form-row">
          <div className="form-group full-width">
            <label htmlFor="comment">Comment</label>
            <textarea
              id="comment"
              value={formData.comment}
              onChange={(e) => handleInputChange('comment', e.target.value)}
              placeholder="Additional notes about this QSO..."
              rows={3}
              disabled={loading}
            />
          </div>
        </div>

        <div className="form-row">
          <div className="form-group checkbox-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={formData.confirmed}
                onChange={(e) => handleInputChange('confirmed', e.target.checked)}
                disabled={loading}
              />
              QSL Confirmed
            </label>
          </div>
        </div>

        <div className="form-actions">
          <button type="button" onClick={onCancel} className="cancel-btn" disabled={loading}>
            Cancel
          </button>
          <button type="submit" className="save-btn" disabled={loading}>
            <Save size={16} />
            {loading ? 'Saving...' : editingContact ? 'Update QSO' : 'Save QSO'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default ContactForm;