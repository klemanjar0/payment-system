import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box, Typography, Card, CardContent, Grid, Chip, Button,
  CircularProgress, Divider,
} from '@mui/material';
import AccountBalanceIcon from '@mui/icons-material/AccountBalance';
import AddCircleOutlineIcon from '@mui/icons-material/AddCircleOutline';
import SwapHorizIcon from '@mui/icons-material/SwapHoriz';
import { useAppDispatch, useAppSelector } from '../store/hooks';
import { getMeThunk } from '../store/slices/authSlice';
import { fetchMyAccounts } from '../store/slices/accountSlice';
import type { Account } from '../types';

export default function DashboardPage(): React.JSX.Element {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { user } = useAppSelector((s) => s.auth);
  const { list: accounts, loading } = useAppSelector((s) => s.accounts);

  useEffect(() => {
    dispatch(getMeThunk());
    dispatch(fetchMyAccounts());
  }, [dispatch]);

  const totalBalance = accounts.reduce((sum, a) => sum + a.balance, 0);
  const totalAvailable = accounts.reduce((sum, a) => sum + a.available, 0);

  return (
    <Box>
      <Typography variant="h4" fontWeight={700} mb={0.5}>Dashboard</Typography>
      {user && (
        <Typography variant="body1" color="text.secondary" mb={3}>
          Welcome back, {user.first_name} {user.last_name}
        </Typography>
      )}

      <Grid container spacing={2} mb={4}>
        <Grid item xs={12} sm={4}>
          <Card>
            <CardContent>
              <Typography color="text.secondary" variant="body2" gutterBottom>Total Balance</Typography>
              <Typography variant="h4" fontWeight={700}>${totalBalance.toFixed(2)}</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={4}>
          <Card>
            <CardContent>
              <Typography color="text.secondary" variant="body2" gutterBottom>Available</Typography>
              <Typography variant="h4" fontWeight={700} color="success.main">
                ${totalAvailable.toFixed(2)}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={4}>
          <Card>
            <CardContent>
              <Typography color="text.secondary" variant="body2" gutterBottom>Accounts</Typography>
              <Typography variant="h4" fontWeight={700}>{accounts.length}</Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Box display="flex" gap={2} mb={4}>
        <Button
          variant="contained" startIcon={<AddCircleOutlineIcon />}
          onClick={() => navigate('/accounts')}
        >
          New Account
        </Button>
        <Button
          variant="outlined" startIcon={<SwapHorizIcon />}
          onClick={() => navigate('/transfer')}
          disabled={accounts.length === 0}
        >
          Send Transfer
        </Button>
      </Box>

      <Divider sx={{ mb: 3 }} />

      <Typography variant="h6" fontWeight={600} mb={2}>Your Accounts</Typography>

      {loading ? (
        <Box display="flex" justifyContent="center" py={4}><CircularProgress /></Box>
      ) : accounts.length === 0 ? (
        <Box textAlign="center" py={6}>
          <AccountBalanceIcon sx={{ fontSize: 56, color: 'text.disabled', mb: 1 }} />
          <Typography color="text.secondary" mb={2}>No accounts yet</Typography>
          <Button variant="contained" onClick={() => navigate('/accounts')}>
            Create Your First Account
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
                    <Typography variant="overline" color="text.secondary" fontWeight={600}>
                      {acc.currency}
                    </Typography>
                    <Chip
                      label={acc.status}
                      size="small"
                      color={acc.status === 'active' ? 'success' : 'default'}
                    />
                  </Box>
                  <Typography variant="h5" fontWeight={700} mb={0.5}>
                    {acc.balance.toFixed(2)}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Available: <strong>{acc.available.toFixed(2)}</strong>
                    {acc.hold_amount > 0 && ` · Hold: ${acc.hold_amount.toFixed(2)}`}
                  </Typography>
                  <Typography variant="caption" color="text.disabled" sx={{ mt: 1, display: 'block' }}>
                    {acc.id}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}
    </Box>
  );
}
