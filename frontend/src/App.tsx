import React, { useState, useEffect } from 'react';
import { Contact, NewContact, SearchFilters, Statistics as StatsType, ImportResult } from './types';
import { qsoApi, downloadFile } from './api';
import ContactList from './components/ContactList.tsx';
import ContactForm from './components/ContactForm.tsx';
import SearchForm from './components/SearchForm.tsx';
import Statistics from './components/Statistics.tsx';
import ImportForm from './components/ImportForm.tsx';
import { 
  Radio, 
  Plus, 
  Search, 
  BarChart3, 
  Download,
  RefreshCw,
  Upload
} from 'lucide-react';
import './App.css';

function App() {
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [filteredContacts, setFilteredContacts] = useState<Contact[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'list' | 'add' | 'search' | 'stats' | 'import'>('list');
  const [editingContact, setEditingContact] = useState<Contact | undefined>(undefined);

  const [searchFilters, setSearchFilters] = useState<SearchFilters>({});
  const [statistics, setStatistics] = useState<StatsType | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);

  // Load contacts on component mount
  useEffect(() => {
    loadContacts();
  }, []);

  const loadContacts = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await qsoApi.getContacts();
      setContacts(data);
      setFilteredContacts(data);
    } catch (err) {
      setError('Failed to load contacts: ' + (err instanceof Error ? err.message : 'Unknown error'));
    } finally {
      setLoading(false);
    }
  };

  const handleAddContact = async (newContact: NewContact) => {
    try {
      if (editingContact) {
        // Update existing contact
        const updatedContact = await qsoApi.updateContact(editingContact.id, newContact);
        setContacts(prev => prev.map(c => c.id === editingContact.id ? updatedContact : c));
        setFilteredContacts(prev => prev.map(c => c.id === editingContact.id ? updatedContact : c));
        setEditingContact(undefined);
      } else {
        // Create new contact
        const contact = await qsoApi.createContact(newContact);
        setContacts(prev => [contact, ...prev]);
        setFilteredContacts(prev => [contact, ...prev]);
      }
      setActiveTab('list');
      return editingContact ? 'updated' : 'created';
    } catch (err) {
      throw new Error(`Failed to ${editingContact ? 'update' : 'add'} contact: ` + (err instanceof Error ? err.message : 'Unknown error'));
    }
  };

  const handleEditContact = (contact: Contact) => {
    setEditingContact(contact);
    setActiveTab('add');
  };

  const handleCancelEdit = () => {
    setEditingContact(undefined);
    setActiveTab('list');
  };

  const handleDeleteContact = async (id: number) => {
    try {
      await qsoApi.deleteContact(id);
      setContacts(prev => prev.filter(c => c.id !== id));
      setFilteredContacts(prev => prev.filter(c => c.id !== id));
    } catch (err) {
      setError('Failed to delete contact: ' + (err instanceof Error ? err.message : 'Unknown error'));
    }
  };

  const handleSearch = async (filters: SearchFilters) => {
    setLoading(true);
    try {
      const results = await qsoApi.searchContacts(filters);
      setFilteredContacts(results);
      setSearchFilters(filters);
    } catch (err) {
      setError('Failed to search contacts: ' + (err instanceof Error ? err.message : 'Unknown error'));
    } finally {
      setLoading(false);
    }
  };

  const handleFiltersChange = (filters: SearchFilters) => {
    setSearchFilters(filters);
    handleSearch(filters);
  };

  const handleClearFilters = () => {
    setSearchFilters({});
    setFilteredContacts(contacts);
  };

  const loadStatistics = async () => {
    setStatsLoading(true);
    try {
      const stats = await qsoApi.getStatistics();
      setStatistics(stats);
    } catch (err) {
      setError('Failed to load statistics: ' + (err instanceof Error ? err.message : 'Unknown error'));
    } finally {
      setStatsLoading(false);
    }
  };

  // Load statistics when switching to stats tab
  useEffect(() => {
    if (activeTab === 'stats' && !statistics) {
      loadStatistics();
    }
  }, [activeTab, statistics]);

  const handleExportADIF = async () => {
    try {
      const blob = await qsoApi.exportADIF();
      const filename = `goqso_export_${new Date().toISOString().split('T')[0]}.adi`;
      downloadFile(blob, filename);
    } catch (err) {
      setError('Failed to export ADIF: ' + (err instanceof Error ? err.message : 'Unknown error'));
    }
  };

  const handleImportComplete = (result: ImportResult) => {
    if (result.success) {
      setError(null);
      // Refresh contacts after successful import
      loadContacts();
      // Show success message
      alert(`Import completed successfully!\n\nImported: ${result.imported_count} contacts\nSkipped: ${result.skipped_count} contacts\nErrors: ${result.error_count} contacts`);
      // Switch back to list view
      setActiveTab('list');
    } else {
      setError(`Import failed: ${result.message}\n${result.errors.join('\n')}`);
    }
  };

  const TabButton = ({ 
    tab, 
    icon: Icon, 
    label, 
    isActive 
  }: { 
    tab: typeof activeTab, 
    icon: React.ComponentType<any>, 
    label: string, 
    isActive: boolean 
  }) => (
    <button
      onClick={() => {
        setActiveTab(tab);
        // Clear editing state when navigating to a different tab
        if (tab !== 'add' && editingContact) {
          setEditingContact(undefined);
        }
      }}
      className={`sidebar-button ${isActive ? 'active' : ''}`}
    >
      <Icon size={20} />
      <span>{label}</span>
    </button>
  );

  return (
    <div className="app">
      <header className="header">
        <div className="header-content">
          <div className="logo">
            <Radio size={32} />
            <h1>GoQSO</h1>
          </div>
          <div className="header-actions">
            <button onClick={loadContacts} className="refresh-btn" disabled={loading}>
              <RefreshCw size={16} className={loading ? 'spinning' : ''} />
              Refresh
            </button>
            <button onClick={handleExportADIF} className="export-btn">
              <Download size={16} />
              Export ADIF
            </button>
          </div>
        </div>
      </header>

      <div className="app-body">
        <nav className="sidebar">
          <div className="sidebar-nav">
            <TabButton tab="list" icon={Radio} label="QSO Log" isActive={activeTab === 'list'} />
            <TabButton tab="add" icon={Plus} label={editingContact ? "Edit QSO" : "Add QSO"} isActive={activeTab === 'add'} />
            <TabButton tab="search" icon={Search} label="Search" isActive={activeTab === 'search'} />
            <TabButton tab="stats" icon={BarChart3} label="Statistics" isActive={activeTab === 'stats'} />
            <TabButton tab="import" icon={Upload} label="Import" isActive={activeTab === 'import'} />
          </div>
        </nav>

        <main className="main-content">
          {error && (
            <div className="error-message">
              {error}
              <button onClick={() => setError(null)}>×</button>
            </div>
          )}

          {activeTab === 'list' && (
            <ContactList 
              contacts={filteredContacts} 
              loading={loading}
              onDelete={handleDeleteContact}
              onEdit={handleEditContact}
            />
          )}

          {activeTab === 'add' && (
            <ContactForm 
              onSave={handleAddContact} 
              onCancel={handleCancelEdit}
              editingContact={editingContact}
            />
          )}

          {activeTab === 'search' && (
            <div>
              <SearchForm 
                filters={searchFilters}
                onFiltersChange={handleFiltersChange}
                onClearFilters={handleClearFilters}
                totalContacts={contacts.length}
                filteredContacts={filteredContacts.length}
              />
              <ContactList 
                contacts={filteredContacts} 
                loading={loading}
                onDelete={handleDeleteContact}
                onEdit={handleEditContact}
              />
            </div>
          )}

          {activeTab === 'stats' && (
            <Statistics 
              statistics={statistics || {
                total_qsos: 0,
                unique_callsigns: 0,
                unique_countries: 0,
                confirmed_qsos: 0,
                qsos_by_band: {},
                qsos_by_mode: {},
                qsos_by_country: {},
                bands_worked: {},
                modes_used: {},
                countries_worked: {},
                date_range: { earliest: '', latest: '' }
              }}
              loading={statsLoading}
            />
          )}

          {activeTab === 'import' && (
            <ImportForm onImportComplete={handleImportComplete} />
          )}
        </main>
      </div>
    </div>
  );
}

export default App;