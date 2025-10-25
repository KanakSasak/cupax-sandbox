import axios from 'axios';
import type { Analysis } from '../types';
import type { Whitelist, WhitelistCreate } from '../types/whitelist';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const uploadFile = async (
  file: File,
  zipPassword?: string
): Promise<{ analysis_id: string; message: string }> => {
  const formData = new FormData();
  formData.append('file', file);

  // Add zip password if provided
  if (zipPassword) {
    formData.append('zip_password', zipPassword);
  }

  const response = await axios.post(`${API_BASE_URL}/analyze`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });

  return response.data;
};

export const getAnalyses = async (): Promise<Analysis[]> => {
  const response = await api.get('/analyses');
  return response.data;
};

export const getAnalysisById = async (id: string): Promise<Analysis> => {
  const response = await api.get(`/analyses/${id}`);
  return response.data;
};

// Whitelist API functions
export const getWhitelists = async (): Promise<Whitelist[]> => {
  const response = await api.get('/whitelists');
  return response.data;
};

export const getWhitelistById = async (id: string): Promise<Whitelist> => {
  const response = await api.get(`/whitelists/${id}`);
  return response.data;
};

export const createWhitelist = async (data: WhitelistCreate): Promise<Whitelist> => {
  const response = await api.post('/whitelists', data);
  return response.data;
};

export const updateWhitelist = async (id: string, data: Partial<WhitelistCreate>): Promise<Whitelist> => {
  const response = await api.put(`/whitelists/${id}`, data);
  return response.data;
};

export const deleteWhitelist = async (id: string): Promise<void> => {
  await api.delete(`/whitelists/${id}`);
};

export default api;
