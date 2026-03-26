import React, { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Box, Typography, Card, CardContent, Chip, Button, CircularProgress,
  Table, TableHead, TableRow, TableCell, TableBody, TableContainer,
  Paper, Alert, Divider,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import SwapHorizIcon from '@mui/icons-material/SwapHoriz';
import { useAppDispatch, useAppSelector } from '../store/hooks';
import { fetchAccount, clearCurrentAccount } from '../store/slices/accountSlice';
import { fetchTransactionsByAccount } from '../store/slices/transactionSlice';
import type { ChipProps } from '@mui/material';

type StatusColor = ChipProps['color'];

const STATUS_COLOR: Record<string, StatusColor> = {
  completed: 'success',
  failed: 'error',
  pending: 'warning',
  processing: 'info',
  reversed: 'default',
};

export default function AccountDetailPage(): React.JSX.Element {
  const { id } = useParams<{ id: string }>();
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { current: account, loading: accLoading } = useAppSelector((s) => s.accounts);
  const { list: txs, total, loading: txLoading, error: txError } = useAppSelector((s) => s.transactions);

  useEffect(() => {
    if (id) {
      dispatch(fetchAccount(id));
      dispatch(fetchTransactionsByAccount({ accountId: id, limit: 50, offset: 0 }));
    }
    return () => { dispatch(clearCurrentAccount()); };
  }, [id, dispatch]);

  if (accLoading || !account) {
    return <Box display="flex" justifyContent="center" mt={8}><CircularProgress /></Box>;
  }

  return (
    <Box>
      <Button startIcon={<ArrowBackIcon />} onClick={() => navigate('/accounts')} sx={{ mb: 2 }}>
        Back to Accounts
      </Button>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Box display="flex" alignItems="center" gap={2} mb={1}>
            <Typography variant="h5" fontWeight={700}>{account.currency} Account</Typography>
            <Chip
              label={account.status}
              color={account.status === 'active' ? 'success' : 'default'}
            />
          </Box>
          <Typography variant="caption" color="text.disabled" sx={{ display: 'block', mb: 2 }}>
            {account.id}
          </Typography>

          <Box display="flex" gap={5}>
            <Box>
              <Typography color="text.secondary" variant="body2">Balance</Typography>
              <Typography variant="h4" fontWeight={700}>{account.balance.toFixed(2)}</Typography>
            </Box>
            <Box>
              <Typography color="text.secondary" variant="body2">Available</Typography>
              <Typography variant="h4" fontWeight={700} color="success.main">
                {account.available.toFixed(2)}
              </Typography>
            </Box>
            <Box>
              <Typography color="text.secondary" variant="body2">On Hold</Typography>
              <Typography variant="h4" fontWeight={700} color="warning.main">
                {account.hold_amount.toFixed(2)}
              </Typography>
            </Box>
          </Box>
        </CardContent>
      </Card>

      <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
        <Typography variant="h6" fontWeight={600}>
          Transactions {total > 0 && `(${total})`}
        </Typography>
        <Button
          variant="outlined" startIcon={<SwapHorizIcon />}
          onClick={() => navigate('/transfer')}
        >
          New Transfer
        </Button>
      </Box>

      {txError && <Alert severity="error" sx={{ mb: 2 }}>{txError}</Alert>}

      {txLoading ? (
        <Box display="flex" justifyContent="center" py={4}><CircularProgress /></Box>
      ) : txs.length === 0 ? (
        <Box textAlign="center" py={4}>
          <Typography color="text.secondary">No transactions yet</Typography>
        </Box>
      ) : (
        <TableContainer component={Paper}>
          <Table size="small">
            <TableHead>
              <TableRow sx={{ bgcolor: 'grey.50' }}>
                <TableCell><strong>Date</strong></TableCell>
                <TableCell><strong>From</strong></TableCell>
                <TableCell><strong>To</strong></TableCell>
                <TableCell align="right"><strong>Amount</strong></TableCell>
                <TableCell><strong>Status</strong></TableCell>
                <TableCell><strong>Description</strong></TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {txs.map((tx) => {
                const isIncoming = tx.to_account_id === id;
                return (
                  <TableRow key={tx.id} hover>
                    <TableCell sx={{ whiteSpace: 'nowrap' }}>
                      {new Date(tx.created_at).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <Typography variant="caption" title={tx.from_account_id}>
                        {tx.from_account_id === id
                          ? <em>This account</em>
                          : `${tx.from_account_id.slice(0, 8)}…`}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography variant="caption" title={tx.to_account_id}>
                        {tx.to_account_id === id
                          ? <em>This account</em>
                          : `${tx.to_account_id.slice(0, 8)}…`}
                      </Typography>
                    </TableCell>
                    <TableCell align="right">
                      <Typography
                        fontWeight={700}
                        color={isIncoming ? 'success.main' : 'error.main'}
                      >
                        {isIncoming ? '+' : '−'}{tx.amount.toFixed(2)} {tx.currency}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={tx.status}
                        size="small"
                        color={STATUS_COLOR[tx.status] ?? 'default'}
                      />
                    </TableCell>
                    <TableCell>{tx.description || '—'}</TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {txs.length > 0 && (
        <Box mt={1}>
          <Divider />
          <Typography variant="caption" color="text.disabled" sx={{ mt: 1, display: 'block' }}>
            Showing {txs.length} of {total} transactions
          </Typography>
        </Box>
      )}
    </Box>
  );
}
