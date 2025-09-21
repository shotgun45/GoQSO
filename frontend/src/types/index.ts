export interface Contact {
  id: number;
  callsign: string;
  contact_date: string; // ISO date string
  time_on: string;
  time_off: string;
  frequency: number;
  band: string;
  mode: string;
  rst_sent: string;
  rst_received: string;
  operator_name: string;
  qth: string;
  country: string;
  grid_square: string;
  power_watts: number;
  comment: string;
  confirmed: boolean;
  created_at: string;
  updated_at: string;
}

export interface NewContact {
  callsign: string;
  operator_name: string;
  contact_date: string;
  time_on: string;
  time_off: string;
  frequency: number;
  band: string;
  mode: string;
  power_watts: number;
  rst_sent: string;
  rst_received: string;
  qth: string;
  country: string;
  grid_square: string;
  comment: string;
  confirmed: boolean;
}

export interface SearchFilters {
  search?: string;
  callsign?: string;
  band?: string;
  mode?: string;
  country?: string;
  dateFrom?: string;
  dateTo?: string;
  date_from?: string;
  date_to?: string;
  freq_min?: number;
  freq_max?: number;
  confirmed?: boolean;
}

export interface Statistics {
  total_qsos: number;
  unique_callsigns: number;
  unique_countries: number;
  confirmed_qsos: number;
  qsos_by_band: Record<string, number>;
  qsos_by_mode: Record<string, number>;
  qsos_by_country: Record<string, number>;
  bands_worked: Record<string, number>;
  modes_used: Record<string, number>;
  countries_worked: Record<string, number>;
  date_range: {
    earliest: string;
    latest: string;
  };
}

export interface ImportResult {
  success: boolean;
  imported_count: number;
  skipped_count: number;
  error_count: number;
  errors: string[];
  message: string;
}

export interface ImportOptions {
  file_type: 'adif' | 'lotw';
  merge_duplicates: boolean;
  update_existing: boolean;
}

export interface LotwCredentials {
  username: string;
  password: string;
  start_date?: string;
  end_date?: string;
}