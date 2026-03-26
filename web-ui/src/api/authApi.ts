import type { AxiosResponse } from 'axios';
import { userApi } from './axios';
import type { LoginResponse, RegisterResponse, RefreshTokenResponse, User } from '../types';

export interface RegisterPayload {
  email: string;
  phone: string;
  password: string;
  first_name: string;
  last_name: string;
}

export const login = (
  email: string,
  password: string
): Promise<AxiosResponse<LoginResponse>> =>
  userApi.post('/v1/auth/login', { email, password, device_info: 'web-ui' });

export const register = (
  data: RegisterPayload
): Promise<AxiosResponse<RegisterResponse>> =>
  userApi.post('/v1/users', data);

export const logout = (
  refreshToken: string | null
): Promise<AxiosResponse<void>> =>
  userApi.post('/v1/auth/logout', { refresh_token: refreshToken });

export const getMe = (): Promise<AxiosResponse<User>> =>
  userApi.get('/v1/users/me');

export const changePassword = (
  userId: string,
  oldPassword: string,
  newPassword: string
): Promise<AxiosResponse<void>> =>
  userApi.post(`/v1/users/${userId}/change-password`, {
    old_password: oldPassword,
    new_password: newPassword,
  });

export const refreshToken = (
  token: string,
): Promise<AxiosResponse<RefreshTokenResponse>> =>
  userApi.post('/v1/auth/refresh', { refresh_token: token, device_info: 'web-ui' });
