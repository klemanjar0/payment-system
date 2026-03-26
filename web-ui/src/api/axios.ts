import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios';
import type { RefreshTokenResponse } from '../types';

interface RetryConfig extends InternalAxiosRequestConfig {
  _retry?: boolean;
}

function createInstance(baseURL: string): AxiosInstance {
  const instance = axios.create({ baseURL });

  instance.interceptors.request.use((config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  });

  instance.interceptors.response.use(
    (res) => res,
    async (error) => {
      const original: RetryConfig = error.config;
      if (error.response?.status === 401 && !original._retry) {
        original._retry = true;
        try {
          const refreshToken = localStorage.getItem('refresh_token');
          const { data } = await axios.post<RefreshTokenResponse>(
            '/user-api/v1/auth/refresh',
            { refresh_token: refreshToken, device_info: 'web-ui' }
          );
          localStorage.setItem('access_token', data.access_token);
          localStorage.setItem('refresh_token', data.refresh_token);
          original.headers.Authorization = `Bearer ${data.access_token}`;
          return instance(original);
        } catch {
          localStorage.clear();
          window.location.href = '/login';
        }
      }
      return Promise.reject(error);
    }
  );

  return instance;
}

export const userApi = createInstance('/user-api');
export const accountApi = createInstance('/account-api');
export const transactionApi = createInstance('/transaction-api');
