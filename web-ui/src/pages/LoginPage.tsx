import React, { useState, useEffect, ChangeEvent, FormEvent } from 'react';
import { useNavigate, Link as RouterLink } from 'react-router-dom';
import {
  Box, Button, TextField, Typography, Alert, Paper, Link, CircularProgress,
} from '@mui/material';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import { useAppDispatch, useAppSelector } from '../store/hooks';
import { loginThunk, clearError } from '../store/slices/authSlice';

interface LoginForm {
  email: string;
  password: string;
}

export default function LoginPage(): React.JSX.Element {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { isAuthenticated, loading, error } = useAppSelector((s) => s.auth);
  const [form, setForm] = useState<LoginForm>({ email: '', password: '' });

  useEffect(() => {
    if (isAuthenticated) navigate('/', { replace: true });
    return () => { dispatch(clearError()); };
  }, [isAuthenticated, navigate, dispatch]);

  const handleChange = (e: ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    dispatch(loginThunk(form));
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: 'grey.100',
      }}
    >
      <Paper elevation={3} sx={{ p: 4, width: 400 }}>
        <Box textAlign="center" mb={3}>
          <LockOutlinedIcon sx={{ fontSize: 40, color: 'primary.main', mb: 1 }} />
          <Typography variant="h5" fontWeight={700}>Sign In</Typography>
          <Typography variant="body2" color="text.secondary">Payment System</Typography>
        </Box>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        <Box component="form" onSubmit={handleSubmit}>
          <TextField
            fullWidth label="Email" name="email" type="email"
            value={form.email} onChange={handleChange} required margin="normal"
            autoComplete="email" autoFocus
          />
          <TextField
            fullWidth label="Password" name="password" type="password"
            value={form.password} onChange={handleChange} required margin="normal"
            autoComplete="current-password"
          />
          <Button
            fullWidth variant="contained" type="submit"
            sx={{ mt: 2, py: 1.2 }} disabled={loading}
          >
            {loading ? <CircularProgress size={22} color="inherit" /> : 'Sign In'}
          </Button>
        </Box>

        <Typography mt={3} textAlign="center" variant="body2">
          Don&apos;t have an account?{' '}
          <Link component={RouterLink} to="/register" fontWeight={600}>
            Register
          </Link>
        </Typography>
      </Paper>
    </Box>
  );
}
