import React, { useState, useCallback } from 'react';
import { SearchFilters } from '../types';
import { Search, X } from 'lucide-react';

interface SearchFormProps {
  filters: SearchFilters;
  onFiltersChange: (filters: SearchFilters) => void;
  onClearFilters: () => void;
  totalContacts: number;
  filteredContacts: number;
}

const SearchForm: React.FC<SearchFormProps> = ({
  filters,
  onFiltersChange,
  onClearFilters,
  totalContacts,
  filteredContacts,
}) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const bands = ['160m', '80m', '40m', '30m', '20m', '17m', '15m', '12m', '10m', '6m', '4m', '2m', '70cm'];
  const modes = ['SSB', 'CW', 'FT8', 'FT4', 'PSK31', 'RTTY', 'AM', 'FM', 'DIGITAL'];

  const handleInputChange = useCallback((field: keyof SearchFilters, value: string | boolean) => {
    onFiltersChange({
      ...filters,
      [field]: value,
    });
  }, [filters, onFiltersChange]);

  const hasActiveFilters = Object.values(filters).some(value => 
    typeof value === 'string' ? value.trim() !== '' : value === true
  );

  const handleClearFilters = () => {
    onClearFilters();
    setIsExpanded(false);
  };

  return (
    <div className="search-form">
      <div className="search-header">
        <div className="search-main">
          <div className="search-input-group">
            <Search size={20} className="search-icon" />
            <input
              type="text"
              placeholder="Search callsigns, names, locations..."
              value={filters.search || ''}
              onChange={(e) => handleInputChange('search', e.target.value)}
              className="search-input"
            />
            {filters.search && (
              <button
                onClick={() => handleInputChange('search', '')}
                className="clear-search-btn"
                title="Clear search"
              >
                <X size={16} />
              </button>
            )}
          </div>

          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className={`expand-btn ${isExpanded ? 'expanded' : ''}`}
            title={isExpanded ? 'Hide filters' : 'Show filters'}
          >
            Advanced
          </button>

          {hasActiveFilters && (
            <button
              onClick={handleClearFilters}
              className="clear-all-btn"
              title="Clear all filters"
            >
              Clear All
            </button>
          )}
        </div>

        <div className="search-results">
          {filteredContacts !== totalContacts ? (
            <span>
              Showing {filteredContacts} of {totalContacts} contacts
            </span>
          ) : (
            <span>{totalContacts} contacts</span>
          )}
        </div>
      </div>

      {isExpanded && (
        <div className="advanced-filters">
          <div className="filter-section">
            <h4>Date Range</h4>
            <div className="filter-row">
              <div className="filter-group">
                <label htmlFor="date_from">From Date</label>
                <input
                  id="date_from"
                  type="date"
                  value={filters.date_from || ''}
                  onChange={(e) => handleInputChange('date_from', e.target.value)}
                />
              </div>
              <div className="filter-group">
                <label htmlFor="date_to">To Date</label>
                <input
                  id="date_to"
                  type="date"
                  value={filters.date_to || ''}
                  onChange={(e) => handleInputChange('date_to', e.target.value)}
                />
              </div>
            </div>
          </div>

          <div className="filter-section">
            <h4>Radio Details</h4>
            <div className="filter-row">
              <div className="filter-group">
                <label htmlFor="band">Band</label>
                <select
                  id="band"
                  value={filters.band || ''}
                  onChange={(e) => handleInputChange('band', e.target.value)}
                >
                  <option value="">All Bands</option>
                  {bands.map(band => (
                    <option key={band} value={band}>{band}</option>
                  ))}
                </select>
              </div>

              <div className="filter-group">
                <label htmlFor="mode">Mode</label>
                <select
                  id="mode"
                  value={filters.mode || ''}
                  onChange={(e) => handleInputChange('mode', e.target.value)}
                >
                  <option value="">All Modes</option>
                  {modes.map(mode => (
                    <option key={mode} value={mode}>{mode}</option>
                  ))}
                </select>
              </div>

              <div className="filter-group">
                <label htmlFor="country">Country</label>
                <input
                  id="country"
                  type="text"
                  placeholder="Filter by country..."
                  value={filters.country || ''}
                  onChange={(e) => handleInputChange('country', e.target.value)}
                />
              </div>
            </div>
          </div>

          <div className="filter-section">
            <h4>Frequency Range</h4>
            <div className="filter-row">
              <div className="filter-group">
                <label htmlFor="freq_min">Min Frequency (MHz)</label>
                <input
                  id="freq_min"
                  type="number"
                  step="0.001"
                  min="0.001"
                  max="999.999"
                  placeholder="14.000"
                  value={filters.freq_min || ''}
                  onChange={(e) => handleInputChange('freq_min', e.target.value)}
                />
              </div>
              <div className="filter-group">
                <label htmlFor="freq_max">Max Frequency (MHz)</label>
                <input
                  id="freq_max"
                  type="number"
                  step="0.001"
                  min="0.001"
                  max="999.999"
                  placeholder="14.350"
                  value={filters.freq_max || ''}
                  onChange={(e) => handleInputChange('freq_max', e.target.value)}
                />
              </div>
            </div>
          </div>

          <div className="filter-section">
            <h4>QSL Status</h4>
            <div className="filter-row">
              <div className="filter-group checkbox-group">
                <label className="checkbox-label">
                  <input
                    type="checkbox"
                    checked={filters.confirmed || false}
                    onChange={(e) => handleInputChange('confirmed', e.target.checked)}
                  />
                  Only QSL Confirmed
                </label>
              </div>
            </div>
          </div>

          <div className="filter-actions">
            <button
              onClick={handleClearFilters}
              className="clear-filters-btn"
            >
              <X size={16} />
              Clear Filters
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default SearchForm;