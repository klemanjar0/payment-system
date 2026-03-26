import type { AxiosResponse } from 'axios';
import { transactionApi } from './axios';
import type { CreateTransferResponse, TransactionListResponse } from '../types';

// amount is in major units (e.g. 10.50) — converted to int64 minor units (cents) before sending
export const createTransfer = (
  fromAccountId: string,
  toAccountId: string,
  amount: number | string,
  currency: string,
  description: string
): Promise<AxiosResponse<CreateTransferResponse>> =>
  transactionApi.post('/v1/transfers', {
    idempotency_key: crypto.randomUUID(),
    from_account_id: fromAccountId,
    to_account_id: toAccountId,
    amount: Math.round(parseFloat(String(amount)) * 100),
    currency,
    description,
  });

export const getTransactionsByAccount = (
  accountId: string,
  limit = 20,
  offset = 0
): Promise<AxiosResponse<TransactionListResponse>> =>
  transactionApi.get(`/v1/accounts/${accountId}/transactions`, {
    params: { limit, offset },
  });
