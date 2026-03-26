import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import type { AxiosError } from 'axios';
import * as accountApi from '../../api/accountApi';
import type { AccountState, Account } from '../../types';

export const fetchMyAccounts = createAsyncThunk<Account[], void, { rejectValue: string }>(
  'accounts/fetchMy',
  async (_, { rejectWithValue }) => {
    try {
      const { data } = await accountApi.getMyAccounts();
      return data.accounts ?? [];
    } catch (e) {
      const err = e as AxiosError<{ error: string }>;
      return rejectWithValue(err.response?.data?.error ?? 'Failed to fetch accounts');
    }
  }
);

export const createAccountThunk = createAsyncThunk<Account, string, { rejectValue: string }>(
  'accounts/create',
  async (currency, { rejectWithValue }) => {
    try {
      const { data } = await accountApi.createAccount(currency);
      return data;
    } catch (e) {
      const err = e as AxiosError<{ error: string }>;
      return rejectWithValue(err.response?.data?.error ?? 'Failed to create account');
    }
  }
);

export const fetchAccount = createAsyncThunk<Account, string, { rejectValue: string }>(
  'accounts/fetchOne',
  async (id, { rejectWithValue }) => {
    try {
      const { data } = await accountApi.getAccount(id);
      return data;
    } catch (e) {
      const err = e as AxiosError<{ error: string }>;
      return rejectWithValue(err.response?.data?.error ?? 'Failed to fetch account');
    }
  }
);

const initialState: AccountState = {
  list: [],
  current: null,
  loading: false,
  error: null,
};

const accountSlice = createSlice({
  name: 'accounts',
  initialState,
  reducers: {
    clearAccountError(state) { state.error = null; },
    clearCurrentAccount(state) { state.current = null; },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchMyAccounts.pending, (state) => { state.loading = true; state.error = null; })
      .addCase(fetchMyAccounts.fulfilled, (state, action) => { state.loading = false; state.list = action.payload; })
      .addCase(fetchMyAccounts.rejected, (state, action) => { state.loading = false; state.error = action.payload ?? null; })
      .addCase(createAccountThunk.pending, (state) => { state.loading = true; state.error = null; })
      .addCase(createAccountThunk.fulfilled, (state, action) => {
        state.loading = false;
        state.list = [...state.list, action.payload];
      })
      .addCase(createAccountThunk.rejected, (state, action) => { state.loading = false; state.error = action.payload ?? null; })
      .addCase(fetchAccount.pending, (state) => { state.loading = true; state.error = null; })
      .addCase(fetchAccount.fulfilled, (state, action) => { state.loading = false; state.current = action.payload; })
      .addCase(fetchAccount.rejected, (state, action) => { state.loading = false; state.error = action.payload ?? null; });
  },
});

export const { clearAccountError, clearCurrentAccount } = accountSlice.actions;
export default accountSlice.reducer;
