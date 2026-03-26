import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import type { AxiosError } from 'axios';
import * as authApi from '../../api/authApi';
import type { AuthState, User, LoginResponse, RegisterResponse } from '../../types';
import type { RegisterPayload } from '../../api/authApi';

interface LoginArgs {
  email: string;
  password: string;
}

export const loginThunk = createAsyncThunk<LoginResponse, LoginArgs, { rejectValue: string }>(
  'auth/login',
  async ({ email, password }, { rejectWithValue }) => {
    try {
      const { data } = await authApi.login(email, password);
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('user_id', data.user_id);
      return data;
    } catch (e) {
      const err = e as AxiosError<{ error: string }>;
      return rejectWithValue(err.response?.data?.error ?? 'Login failed');
    }
  }
);

export const registerThunk = createAsyncThunk<RegisterResponse, RegisterPayload, { rejectValue: string }>(
  'auth/register',
  async (userData, { rejectWithValue }) => {
    try {
      const { data } = await authApi.register(userData);
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('user_id', data.user_id);
      return data;
    } catch (e) {
      const err = e as AxiosError<{ error: string }>;
      return rejectWithValue(err.response?.data?.error ?? 'Registration failed');
    }
  }
);

export const getMeThunk = createAsyncThunk<User, void, { rejectValue: string }>(
  'auth/getMe',
  async (_, { rejectWithValue }) => {
    try {
      const { data } = await authApi.getMe();
      return data;
    } catch (e) {
      const err = e as AxiosError<{ error: string }>;
      return rejectWithValue(err.response?.data?.error ?? 'Failed to get profile');
    }
  }
);

export const logoutThunk = createAsyncThunk<void, void>(
  'auth/logout',
  async () => {
    try {
      const refreshToken = localStorage.getItem('refresh_token');
      await authApi.logout(refreshToken);
    } catch {
      // ignore — always clear local storage
    }
    localStorage.clear();
  }
);

const initialState: AuthState = {
  user: null,
  userId: localStorage.getItem('user_id'),
  isAuthenticated: !!localStorage.getItem('access_token'),
  loading: false,
  error: null,
};

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    clearError(state) {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loginThunk.pending, (state) => { state.loading = true; state.error = null; })
      .addCase(loginThunk.fulfilled, (state, action) => {
        state.loading = false;
        state.isAuthenticated = true;
        state.userId = action.payload.user_id;
      })
      .addCase(loginThunk.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload ?? 'Login failed';
      })
      .addCase(registerThunk.pending, (state) => { state.loading = true; state.error = null; })
      .addCase(registerThunk.fulfilled, (state, action) => {
        state.loading = false;
        state.isAuthenticated = true;
        state.userId = action.payload.user_id;
      })
      .addCase(registerThunk.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload ?? 'Registration failed';
      })
      .addCase(getMeThunk.fulfilled, (state, action) => {
        state.user = action.payload;
      })
      .addCase(logoutThunk.fulfilled, (state) => {
        state.user = null;
        state.userId = null;
        state.isAuthenticated = false;
      });
  },
});

export const { clearError } = authSlice.actions;
export default authSlice.reducer;
