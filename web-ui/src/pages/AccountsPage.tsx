import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box, Typography, Button, Card, CardContent, Grid, Chip, CircularProgress,
  Alert, Dialog, DialogTitle, DialogContent, DialogActions,
  MenuItem, Select, FormControl, InputLabel, SelectChangeEvent,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import { useAppDispatch, useAppSelector } from '../store/hooks';
import { fetchMyAccounts, createAccountThunk, clearAccountError } from '../store/slices/accountSlice';
import type { Account } from '../types';

const CURRENCIES = ['USD', 'EUR', 'GBP', 'UAH', 'PLN'];

export default function AccountsPage(): React.JSX.Element {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { list: accounts, loading, error } = useAppSelector((s) => s.accounts);
  const [open, setOpen] = useState(false);
  const [currency, setCurrency] = useState('USD');

  useEffect(() => {
    dispatch(fetchMyAccounts());
    return () => { dispatch(clearAccountError()); };
  }, [dispatch]);

  const handleCreate = async () => {
    const result = await dispatch(createAccountThunk(currency));
    if (!result.error) {
      setOpen(false);
      setCurrency('USD');
    }
  };

  return (
    <Box>
      <Box display="flex" alignItems="center" justifyContent="space-between" mb={3}>
        <Typography variant="h4" fontWeight={700}>Accounts</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpen(true)}>
          New Account
        </Button>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      {loading ? (
        <Box display="flex" justifyContent="center" py={6}><CircularProgress /></Box>
      ) : accounts.length === 0 ? (
        <Box textAlign="center" py={6}>
          <Typography color="text.secondary" mb={2}>No accounts yet</Typography>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpen(true)}>
            Create Account
          </Button>
        </Box>
      ) : (
        <Grid container spacing={2}>
          {accounts.map((acc: Account) => (
            <Grid item xs={12} sm={6} md={4} key={acc.id}>
              <Card
                sx={{ cursor: 'pointer', transition: '0.2s', '&:hover': { boxShadow: 6 } }}
                onClick={() => navigate(`/accounts/${acc.id}`)}
              >
                <CardContent>
                  <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                    <Typography variant="overline" fontWeight={700} color="primary">
                      {acc.currency}
                    </Typography>
                    <Chip
                      label={acc.status}
                      size="small"
                      color={acc.status === 'active' ? 'success' : 'default'}
                    />
                  </Box>
                  <Typography variant="h4" fontWeight={700} mb={1}>
                    {acc.balance.toFixed(2)}
                  </Typography>
                  <Box display="flex" gap={2}>
                    <Box>
                      <Typography variant="caption" color="text.secondary">Available</Typography>
                      <Typography variant="body2" fontWeight={600} color="success.main">
                        {acc.available.toFixed(2)}
                      </Typography>
                    </Box>
                    <Box>
                      <Typography variant="caption" color="text.secondary">On Hold</Typography>
                      <Typography variant="body2" fontWeight={600} color="warning.main">
                        {acc.hold_amount.toFixed(2)}
                      </Typography>
                    </Box>
                  </Box>
                  <Typography variant="caption" color="text.disabled" sx={{ mt: 1.5, display: 'block' }}>
                    {acc.id}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>Create New Account</DialogTitle>
        <DialogContent>
          <FormControl fullWidth sx={{ mt: 1 }}>
            <InputLabel>Currency</InputLabel>
            <Select
              value={currency}
              label="Currency"
              onChange={(e: SelectChangeEvent) => setCurrency(e.target.value)}
            >
              {CURRENCIES.map((c) => <MenuItem key={c} value={c}>{c}</MenuItem>)}
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleCreate} disabled={loading}>
            {loading ? <CircularProgress size={20} /> : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
