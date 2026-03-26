// ── Auth ────────────────────────────────────────────────────────────────────

export interface LoginResponse {
  user_id: string;
  access_token: string;
  refresh_token: string;
}

export interface RegisterResponse {
  user_id: string;
  email: string;
  access_token: string;
  refresh_token: string;
  created_at: string;
}

export interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
}

export interface User {
  id: string;
  email: string;
  phone: string;
  first_name: string;
  last_name: string;
  status: string;
  kyc_status: string;
  created_at: string;
  updated_at: string;
}

// ── Account ──────────────────────────────────────────────────────────────────

export interface Account {
  id: string;
  user_id: string;
  currency: string;
  balance: number;
  hold_amount: number;
  available: number;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface AccountListResponse {
  accounts: Account[];
}

// ── Transaction ───────────────────────────────────────────────────────────────

export interface Transaction {
  id: string;
  idempotency_key: string;
  from_account_id: string;
  to_account_id: string;
  amount: number;
  currency: string;
  description: string;
  status: string;
  failure_reason?: string;
  created_at: string;
  updated_at: string;
}

export interface TransactionListResponse {
  transactions: Transaction[];
  total: number;
}

export interface CreateTransferResponse {
  transaction_id: string;
  status: string;
}

// ── Redux state ───────────────────────────────────────────────────────────────

export interface AuthState {
  user: User | null;
  userId: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
}

export interface AccountState {
  list: Account[];
  current: Account | null;
  loading: boolean;
  error: string | null;
}

export interface TransactionState {
  list: Transaction[];
  total: number;
  lastTransfer: CreateTransferResponse | null;
  loading: boolean;
  error: string | null;
}

// ── Misc ──────────────────────────────────────────────────────────────────────

export interface ApiError {
  error: string;
  message?: string;
  code: number;
}
