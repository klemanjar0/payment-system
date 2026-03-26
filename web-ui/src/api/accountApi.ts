import type { AxiosResponse } from 'axios';
import { accountApi } from './axios';
import type { Account, AccountListResponse } from '../types';

export const getMyAccounts = (): Promise<AxiosResponse<AccountListResponse>> =>
  accountApi.get('/v1/accounts/me');

export const createAccount = (currency: string): Promise<AxiosResponse<Account>> =>
  accountApi.post('/v1/accounts', { currency });

export const getAccount = (id: string): Promise<AxiosResponse<Account>> =>
  accountApi.get(`/v1/accounts/${id}`);
