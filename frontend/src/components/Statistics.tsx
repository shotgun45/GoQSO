import React from 'react';
import { Statistics as StatsType } from '../types';
import { BarChart3, Radio, Users, Globe } from 'lucide-react';

interface StatisticsProps {
  statistics: StatsType;
  loading: boolean;
}

const Statistics: React.FC<StatisticsProps> = ({ statistics, loading }) => {
  if (loading) {
    return (
      <div className="statistics">
        <div className="loading">
          <div className="spinner"></div>
          <p>Loading statistics...</p>
        </div>
      </div>
    );
  }

  const StatCard: React.FC<{
    title: string;
    value: number | string;
    icon: React.ReactNode;
    subtitle?: string;
  }> = ({ title, value, icon, subtitle }) => (
    <div className="stat-card">
      <div className="stat-icon">{icon}</div>
      <div className="stat-content">
        <h3>{title}</h3>
        <div className="stat-value">{value}</div>
        {subtitle && <div className="stat-subtitle">{subtitle}</div>}
      </div>
    </div>
  );

  const BandChart: React.FC<{ bands: Record<string, number> }> = ({ bands }) => {
    const maxCount = Math.max(...Object.values(bands));
    const sortedBands = Object.entries(bands)
      .sort(([, a], [, b]) => b - a)
      .slice(0, 10); // Show top 10 bands

    return (
      <div className="band-chart">
        <h3>
          <BarChart3 size={20} />
          QSOs by Band
        </h3>
        <div className="chart-bars">
          {sortedBands.map(([band, count]) => (
            <div key={band} className="chart-bar">
              <div className="bar-label">{band}</div>
              <div className="bar-container">
                <div
                  className="bar-fill"
                  style={{
                    width: `${(count / maxCount) * 100}%`,
                  }}
                ></div>
              </div>
              <div className="bar-value">{count}</div>
            </div>
          ))}
        </div>
      </div>
    );
  };

  const ModeChart: React.FC<{ modes: Record<string, number> }> = ({ modes }) => {
    const maxCount = Math.max(...Object.values(modes));
    const sortedModes = Object.entries(modes)
      .sort(([, a], [, b]) => b - a)
      .slice(0, 8); // Show top 8 modes

    return (
      <div className="mode-chart">
        <h3>
          <Radio size={20} />
          QSOs by Mode
        </h3>
        <div className="chart-bars">
          {sortedModes.map(([mode, count]) => (
            <div key={mode} className="chart-bar">
              <div className="bar-label">{mode}</div>
              <div className="bar-container">
                <div
                  className="bar-fill mode-bar"
                  style={{
                    width: `${(count / maxCount) * 100}%`,
                  }}
                ></div>
              </div>
              <div className="bar-value">{count}</div>
            </div>
          ))}
        </div>
      </div>
    );
  };

  const CountryList: React.FC<{ countries: Record<string, number> }> = ({ countries }) => {
    const sortedCountries = Object.entries(countries)
      .sort(([, a], [, b]) => b - a)
      .slice(0, 10); // Show top 10 countries

    return (
      <div className="country-list">
        <h3>
          <Globe size={20} />
          Countries Worked
        </h3>
        <div className="country-items">
          {sortedCountries.map(([country, count]) => (
            <div key={country} className="country-item">
              <span className="country-name">{country || 'Unknown'}</span>
              <span className="country-count">{count}</span>
            </div>
          ))}
        </div>
      </div>
    );
  };

  return (
    <div className="statistics">
      <div className="statistics-header">
        <h2>QSO Statistics</h2>
      </div>

      <div className="stats-overview">
        <StatCard
          title="Total QSOs"
          value={statistics.total_qsos}
          icon={<Radio size={24} />}
          subtitle="All contacts"
        />

        <StatCard
          title="Countries"
          value={statistics.unique_countries}
          icon={<Globe size={24} />}
          subtitle="Worked"
        />

        <StatCard
          title="QSL Confirmed"
          value={statistics.confirmed_qsos}
          icon={<Users size={24} />}
          subtitle={`${Math.round((statistics.confirmed_qsos / statistics.total_qsos) * 100) || 0}% of total`}
        />

        <StatCard
          title="Unique Calls"
          value={statistics.unique_callsigns}
          icon={<BarChart3 size={24} />}
          subtitle="Different stations"
        />
      </div>

      <div className="charts-grid">
        <div className="chart-section">
          <BandChart bands={statistics.qsos_by_band} />
        </div>

        <div className="chart-section">
          <ModeChart modes={statistics.qsos_by_mode} />
        </div>

        <div className="chart-section">
          <CountryList countries={statistics.qsos_by_country} />
        </div>
      </div>

      {statistics.total_qsos > 0 && (
        <div className="additional-stats">
          <div className="stat-row">
            <div className="stat-item">
              <strong>Average QSOs per day:</strong>
              <span>
                {(statistics.total_qsos / Math.max(1, 
                  Math.ceil((Date.now() - new Date('2024-01-01').getTime()) / (1000 * 60 * 60 * 24))
                )).toFixed(1)}
              </span>
            </div>
            <div className="stat-item">
              <strong>QSL Confirmation Rate:</strong>
              <span>
                {Math.round((statistics.confirmed_qsos / statistics.total_qsos) * 100) || 0}%
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Statistics;