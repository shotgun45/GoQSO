import React, { useState } from 'react';
import { Contact, PaginatedResponse } from '../types';
import { Trash2, MapPin, Radio, Edit, Grid3X3, List } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import Pagination from './Pagination';

interface ContactListProps {
  data: PaginatedResponse<Contact> | null;
  loading: boolean;
  onDelete: (id: number) => void;
  onEdit: (contact: Contact) => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
}

const ContactList: React.FC<ContactListProps> = ({ 
  data, 
  loading, 
  onDelete, 
  onEdit, 
  onPageChange, 
  onPageSizeChange 
}) => {
  const [viewMode, setViewMode] = useState<'card' | 'list'>('card');

  // Helper function to format time for display
  const formatTimeDisplay = (timeString: string): string => {
    if (!timeString) return '';
    
    // If it's already in HH:MM or HH:MM:SS format, return as is (show full format)
    if (timeString.includes(':')) {
      return timeString;
    }
    
    // If it's in HHMM format (like "1746"), convert to HH:MM
    if (timeString.length >= 4) {
      const hours = timeString.slice(0, 2);
      const minutes = timeString.slice(2, 4);
      return `${hours}:${minutes}`;
    }
    
    return timeString;
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="spinner"></div>
        <p>Loading contacts...</p>
      </div>
    );
  }

  if (!data || data.items.length === 0) {
    return (
      <div className="empty-state">
        <Radio size={48} className="empty-icon" />
        <h3>No QSOs Found</h3>
        <p>Start logging your amateur radio contacts!</p>
      </div>
    );
  }

  const contacts = data.items;

  const handleDelete = (contact: Contact) => {
    if (window.confirm(`Delete QSO with ${contact.callsign}?`)) {
      onDelete(contact.id);
    }
  };

  return (
    <div className="contact-list">
      <div className="contact-list-header">
        <h2>QSO Log ({data.total_items} contacts)</h2>
        <div className="view-toggle">
          <button
            onClick={() => setViewMode('card')}
            className={`view-toggle-btn ${viewMode === 'card' ? 'active' : ''}`}
            title="Card View"
          >
            <Grid3X3 size={18} />
          </button>
          <button
            onClick={() => setViewMode('list')}
            className={`view-toggle-btn ${viewMode === 'list' ? 'active' : ''}`}
            title="List View"
          >
            <List size={18} />
          </button>
        </div>
      </div>

      {viewMode === 'card' ? (
        <div className="contact-grid">
          {contacts.map((contact: Contact) => (
            <div key={contact.id} className="contact-card">
            <div className="contact-header">
              <div className="callsign-section">
                <h3 className="callsign">
                  <a 
                    href={`https://www.qrz.com/db/${contact.callsign}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="callsign-link"
                  >
                    {contact.callsign}
                  </a>
                </h3>
                {contact.operator_name && (
                  <span className="operator-name">{contact.operator_name}</span>
                )}
              </div>
              <div className="contact-actions">
                <button
                  onClick={() => onEdit(contact)}
                  className="edit-btn"
                  title="Edit QSO"
                >
                  <Edit size={16} />
                </button>
                <button
                  onClick={() => handleDelete(contact)}
                  className="delete-btn"
                  title="Delete QSO"
                >
                  <Trash2 size={16} />
                </button>
              </div>
            </div>

            <div className="contact-details">
              <div className="detail-row">
                <span className="label">Date:</span>
                <span className="value">
                  {new Date(contact.contact_date).toLocaleDateString()}
                </span>
              </div>

              <div className="detail-row">
                <span className="label">Time:</span>
                <span className="value">
                  {formatTimeDisplay(contact.time_on)} - {formatTimeDisplay(contact.time_off)} UTC
                </span>
              </div>

              <div className="detail-row">
                <span className="label">Frequency:</span>
                <span className="value">
                  {contact.frequency.toFixed(3)} MHz ({contact.band})
                </span>
              </div>

              <div className="detail-row">
                <span className="label">Mode:</span>
                <span className="value">{contact.mode}</span>
              </div>

              <div className="detail-row">
                <span className="label">Signal:</span>
                <span className="value">
                  {contact.rst_sent} / {contact.rst_received}
                </span>
              </div>

              {contact.qth && (
                <div className="detail-row">
                  <span className="label">
                    <MapPin size={14} />
                    QTH:
                  </span>
                  <span className="value">{contact.qth}</span>
                </div>
              )}

              {contact.country && (
                <div className="detail-row">
                  <span className="label">Country:</span>
                  <span className="value">{contact.country}</span>
                </div>
              )}

              {contact.grid_square && (
                <div className="detail-row">
                  <span className="label">Grid:</span>
                  <span className="value">{contact.grid_square}</span>
                </div>
              )}

              {contact.power_watts > 0 && (
                <div className="detail-row">
                  <span className="label">Power:</span>
                  <span className="value">{contact.power_watts}W</span>
                </div>
              )}

              {contact.comment && (
                <div className="detail-row">
                  <span className="label">Comment:</span>
                  <span className="value comment">{contact.comment}</span>
                </div>
              )}
            </div>

            <div className="contact-footer">
              <div className="confirmation-status">
                {contact.confirmed && (
                  <span className="confirmed">✓ QSL Confirmed</span>
                )}
              </div>
              <div className="timestamp">
                Added {formatDistanceToNow(new Date(contact.created_at))} ago
              </div>
            </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="contact-table-container">
          <table className="contact-table">
            <thead>
              <tr>
                <th>Callsign</th>
                <th>Date</th>
                <th>Time (UTC)</th>
                <th>Freq/Band</th>
                <th>Mode</th>
                <th>RST</th>
                <th>QTH</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {contacts.map((contact: Contact) => (
                <tr key={contact.id} className="contact-row">
                  <td className="callsign-cell">
                    <div className="callsign-info">
                      <a 
                        href={`https://www.qrz.com/db/${contact.callsign}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="callsign-link"
                      >
                        {contact.callsign}
                      </a>
                      {contact.operator_name && (
                        <div className="operator-name-small">{contact.operator_name}</div>
                      )}
                    </div>
                  </td>
                  <td>{new Date(contact.contact_date).toLocaleDateString()}</td>
                  <td>
                    {formatTimeDisplay(contact.time_on)}
                    {contact.time_off && ` - ${formatTimeDisplay(contact.time_off)}`}
                  </td>
                  <td>
                    <div className="freq-info">
                      <div>{contact.frequency.toFixed(3)} MHz</div>
                      <div className="band-small">{contact.band}</div>
                    </div>
                  </td>
                  <td>{contact.mode}</td>
                  <td>{contact.rst_sent} / {contact.rst_received}</td>
                  <td>
                    <div className="location-info">
                      {contact.qth && <div>{contact.qth}</div>}
                      {contact.country && <div className="country-small">{contact.country}</div>}
                      {contact.grid_square && <div className="grid-small">{contact.grid_square}</div>}
                    </div>
                  </td>
                  <td>
                    <div className="table-actions">
                      <button
                        onClick={() => onEdit(contact)}
                        className="edit-btn-small"
                        title="Edit QSO"
                      >
                        <Edit size={14} />
                      </button>
                      <button
                        onClick={() => handleDelete(contact)}
                        className="delete-btn-small"
                        title="Delete QSO"
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Pagination
        currentPage={data.page}
        totalPages={data.total_pages}
        pageSize={data.page_size}
        totalItems={data.total_items}
        onPageChange={onPageChange}
        onPageSizeChange={onPageSizeChange}
        loading={loading}
      />
    </div>
  );
};

export default ContactList;