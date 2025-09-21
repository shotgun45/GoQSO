import axios, { AxiosResponse } from 'axios';
import { Contact, NewContact, SearchFilters, Statistics, PaginatedResponse } from '../types';

// Backend API response wrapper
interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

// Backend Contact structure (what we get from API)
interface BackendContact {
  ID: number;
  Callsign: string;
  Date: string;
  TimeOn: string;
  TimeOff: string;
  Frequency: number;
  Band: string;
  Mode: string;
  RSTSent: string;
  RSTReceived: string;
  Name: string;
  QTH: string;
  Country: string;
  Grid: string;
  Power: number;
  Comment: string;
  Confirmed: boolean;
  CreatedAt: string;
  UpdatedAt: string;
}

// Transform backend contact to frontend format
const transformContact = (backendContact: BackendContact): Contact => ({
  id: backendContact.ID,
  callsign: backendContact.Callsign,
  contact_date: backendContact.Date,
  time_on: backendContact.TimeOn,
  time_off: backendContact.TimeOff,
  frequency: backendContact.Frequency,
  band: backendContact.Band,
  mode: backendContact.Mode,
  rst_sent: backendContact.RSTSent,
  rst_received: backendContact.RSTReceived,
  operator_name: backendContact.Name,
  qth: backendContact.QTH,
  country: backendContact.Country,
  grid_square: backendContact.Grid,
  power_watts: backendContact.Power,
  comment: backendContact.Comment,
  confirmed: backendContact.Confirmed,
  created_at: backendContact.CreatedAt,
  updated_at: backendContact.UpdatedAt,
});

const API_BASE_URL = '/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const qsoApi = {
  // Get all contacts (with pagination)
  getContacts: async (page: number = 1, pageSize: number = 20): Promise<PaginatedResponse<Contact>> => {
    const response: AxiosResponse<APIResponse<PaginatedResponse<BackendContact>>> = await api.get(`/contacts?page=${page}&page_size=${pageSize}`);
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to get contacts');
    }
    const data = response.data.data!;
    return {
      items: data.items.map(transformContact),
      page: data.page,
      page_size: data.page_size,
      total_items: data.total_items,
      total_pages: data.total_pages,
    };
  },

  // Get all contacts (legacy method for backward compatibility)
  getAllContacts: async (): Promise<Contact[]> => {
    // Get first 1000 items to maintain compatibility
    const response = await qsoApi.getContacts(1, 1000);
    return response.items;
  },

  // Get contact by ID
  getContact: async (id: number): Promise<Contact> => {
    const response: AxiosResponse<APIResponse<BackendContact>> = await api.get(`/contacts/${id}`);
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to get contact');
    }
    return transformContact(response.data.data!);
  },

  // Create new contact
  createContact: async (contact: NewContact): Promise<Contact> => {
    const response: AxiosResponse<APIResponse<BackendContact>> = await api.post('/contacts', contact);
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to create contact');
    }
    return transformContact(response.data.data!);
  },

  // Update contact
  updateContact: async (id: number, contact: Partial<NewContact>): Promise<Contact> => {
    const response: AxiosResponse<APIResponse<BackendContact>> = await api.put(`/contacts/${id}`, contact);
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to update contact');
    }
    return transformContact(response.data.data!);
  },

  // Delete contact
  deleteContact: async (id: number): Promise<void> => {
    const response: AxiosResponse<APIResponse<any>> = await api.delete(`/contacts/${id}`);
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to delete contact');
    }
  },

  // Search contacts (with pagination)
  searchContacts: async (filters: SearchFilters): Promise<PaginatedResponse<Contact>> => {
    const response: AxiosResponse<APIResponse<PaginatedResponse<BackendContact>>> = await api.post('/contacts/search', filters);
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to search contacts');
    }
    const data = response.data.data!;
    return {
      items: data.items.map(transformContact),
      page: data.page,
      page_size: data.page_size,
      total_items: data.total_items,
      total_pages: data.total_pages,
    };
  },

  // Legacy search method for backward compatibility
  searchContactsLegacy: async (filters: SearchFilters): Promise<Contact[]> => {
    const response = await qsoApi.searchContacts({ ...filters, page: 1, page_size: 1000 });
    return response.items;
  },

  // Get statistics
  getStatistics: async (): Promise<Statistics> => {
    const response: AxiosResponse<APIResponse<Statistics>> = await api.get('/statistics');
    if (!response.data.success) {
      throw new Error(response.data.error || 'Failed to get statistics');
    }
    return response.data.data!;
  },

  // Export ADIF
  exportADIF: async (): Promise<Blob> => {
    const response = await api.get('/contacts/export', {
      responseType: 'blob',
    });
    return response.data;
  },
};

// Utility function to download file
export const downloadFile = (blob: Blob, filename: string) => {
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
};

export default api;