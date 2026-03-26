import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import type { AxiosError } from 'axios';
import * as transactionApi from '../../api/transactionApi';
import type { TransactionState, Transaction, CreateTransferResponse, TransactionListResponse } from '../../types';

interface CreateTransferArgs {
  fromAccountId: string;
  toAccountId: string;
  amount: number | string;
  currency: string;
  description: string;
}

interface FetchByAccountArgs {
  accountId: string;
  limit?: number;
  offset?: number;
}

export const createTransferThunk = createAsyncThunk<
  CreateTransferResponse,
  CreateTransferArgs,
  { rejectValue: string }
>(
  'transactions/createTransfer',
  async ({ fromAccountId, toAccountId, amount, currency, description }, { rejectWithValue }) => {
    try {
      const { data } = await transactionApi.createTransfer(
        fromAccountId, toAccountId, amount, currency, description
      );
      return data;
    } catch (e) {
      const err = e as AxiosError<{ error: string }>;
      return rejectWithValue(err.response?.data?.error ?? 'Transfer failed');
    }
  }
);

export const fetchTransactionsByAccount = createAsyncThunk<
  TransactionListResponse,
  FetchByAccountArgs,
  { rejectValue: string }
>(
  'transactions/fetchByAccount',
  async ({ accountId, limit = 20, offset = 0 }, { rejectWithValue }) => {
    try {
      const { data } = await transactionApi.getTransactionsByAccount(accountId, limit, offset);
      return data;
    } catch (e) {
      const err = e as AxiosError<{ error: string }>;
      return rejectWithValue(err.response?.data?.error ?? 'Failed to fetch transactions');
    }
  }
);

const initialState: TransactionState = {
  list: [],
  total: 0,
  lastTransfer: null,
  loading: false,
  error: null,
};

const transactionSlice = createSlice({
  name: 'transactions',
  initialState,
  reducers: {
    clearTransactionError(state) { state.error = null; },
    clearLastTransfer(state) { state.lastTransfer = null; },
  },
  extraReducers: (builder) => {
    builder
      .addCase(createTransferThunk.pending, (state) => {
        state.loading = true;
        state.error = null;
        state.lastTransfer = null;
      })
      .addCase(createTransferThunk.fulfilled, (state, action) => {
        state.loading = false;
        state.lastTransfer = action.payload;
      })
      .addCase(createTransferThunk.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload ?? null;
      })
      .addCase(fetchTransactionsByAccount.pending, (state) => { state.loading = true; state.error = null; })
      .addCase(fetchTransactionsByAccount.fulfilled, (state, action) => {
        state.loading = false;
        state.list = action.payload.transactions ?? [];
        state.total = action.payload.total ?? 0;
      })
      .addCase(fetchTransactionsByAccount.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload ?? null;
      });
  },
});

export const { clearTransactionError, clearLastTransfer } = transactionSlice.actions;
export default transactionSlice.reducer;
