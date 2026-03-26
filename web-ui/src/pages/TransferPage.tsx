import React, { useEffect, useState, ChangeEvent, FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box, Typography, Card, CardContent, TextField, Button, Alert, CircularProgress,
  MenuItem, Select, FormControl, InputLabel, Divider, SelectChangeEvent,
} from '@mui/material';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import { useAppDispatch, useAppSelector } from '../store/hooks';
import { createTransferThunk, clearLastTransfer, clearTransactionError } from '../store/slices/transactionSlice';
import { fetchMyAccounts } from '../store/slices/accountSlice';

interface TransferForm {
  from_account_id: string;
  to_account_id: string;
  amount: string;
  description: string;
}

export default function TransferPage(): React.JSX.Element {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { list: accounts } = useAppSelector((s) => s.accounts);
  const { loading, error, lastTransfer } = useAppSelector((s) => s.transactions);
  const [form, setForm] = useState<TransferForm>({
    from_account_id: '',
    to_account_id: '',
    amount: '',
    description: '',
  });

  useEffect(() => {
    dispatch(fetchMyAccounts());
    return () => {
      dispatch(clearLastTransfer());
      dispatch(clearTransactionError());
    };
  }, [dispatch]);

  const handleChange = (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));

  const handleSelectChange = (e: SelectChangeEvent) =>
    setForm((f) => ({ ...f, from_account_id: e.target.value }));

  const selectedAccount = accounts.find((a) => a.id === form.from_account_id);

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    dispatch(createTransferThunk({
      fromAccountId: form.from_account_id,
      toAccountId: form.to_account_id,
      amount: form.amount,
      currency: selectedAccount?.currency ?? 'USD',
      description: form.description,
    }));
  };

  const handleReset = () => {
    setForm({ from_account_id: '', to_account_id: '', amount: '', description: '' });
    dispatch(clearLastTransfer());
    dispatch(clearTransactionError());
  };

  if (lastTransfer) {
    return (
      <Box>
        <Typography variant="h4" fontWeight={700} mb={3}>Transfer</Typography>
        <Card sx={{ maxWidth: 520 }}>
          <CardContent sx={{ textAlign: 'center', py: 5 }}>
            <CheckCircleIcon sx={{ fontSize: 72, color: 'success.main', mb: 2 }} />
            <Typography variant="h5" fontWeight={700} mb={1}>Transfer Initiated</Typography>
            <Typography color="text.secondary" mb={1}>Transaction ID:</Typography>
            <Typography
              variant="body2" fontFamily="monospace"
              bgcolor="grey.100" px={2} py={0.5} borderRadius={1} mb={2}
            >
              {lastTransfer.transaction_id}
            </Typography>
            <Typography variant="body1" mb={3}>
              Status:{' '}
              <strong style={{ textTransform: 'capitalize' }}>{lastTransfer.status}</strong>
            </Typography>
            <Box display="flex" gap={2} justifyContent="center">
              <Button variant="contained" onClick={handleReset}>New Transfer</Button>
              <Button variant="outlined" onClick={() => navigate('/accounts')}>
                View Accounts
              </Button>
            </Box>
          </CardContent>
        </Card>
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h4" fontWeight={700} mb={3}>Transfer</Typography>
      <Card sx={{ maxWidth: 520 }}>
        <CardContent sx={{ p: 3 }}>
          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

          {accounts.length === 0 ? (
            <Alert severity="info">
              You need at least one account to make a transfer.{' '}
              <Button size="small" onClick={() => navigate('/accounts')}>Create Account</Button>
            </Alert>
          ) : (
            <Box component="form" onSubmit={handleSubmit}>
              <FormControl fullWidth margin="normal" required>
                <InputLabel>From Account</InputLabel>
                <Select
                  value={form.from_account_id}
                  label="From Account"
                  onChange={handleSelectChange}
                >
                  {accounts.map((acc) => (
                    <MenuItem key={acc.id} value={acc.id}>
                      {acc.currency} — available {acc.available.toFixed(2)}
                      <Typography variant="caption" color="text.disabled" ml={1}>
                        ({acc.id.slice(0, 8)}…)
                      </Typography>
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              {selectedAccount && (
                <Typography variant="caption" color="text.secondary" sx={{ ml: 1 }}>
                  Available: {selectedAccount.available.toFixed(2)} {selectedAccount.currency}
                </Typography>
              )}

              <TextField
                fullWidth
                label="To Account ID"
                name="to_account_id"
                value={form.to_account_id}
                onChange={handleChange}
                required
                margin="normal"
                placeholder="Recipient account UUID"
                inputProps={{ style: { fontFamily: 'monospace' } }}
              />

              <TextField
                fullWidth
                label="Amount"
                name="amount"
                type="number"
                value={form.amount}
                onChange={handleChange}
                required
                margin="normal"
                inputProps={{ min: 0.01, step: 0.01 }}
                helperText={selectedAccount ? `Currency: ${selectedAccount.currency}` : ''}
              />

              <TextField
                fullWidth
                label="Description"
                name="description"
                value={form.description}
                onChange={handleChange}
                margin="normal"
                placeholder="Optional note"
                multiline
                rows={2}
              />

              <Divider sx={{ my: 2 }} />

              <Button
                fullWidth variant="contained" type="submit"
                size="large" disabled={loading}
              >
                {loading ? <CircularProgress size={24} color="inherit" /> : 'Send Transfer'}
              </Button>
            </Box>
          )}
        </CardContent>
      </Card>
    </Box>
  );
}
