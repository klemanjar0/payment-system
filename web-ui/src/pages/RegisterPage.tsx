import React, { useState, useEffect, ChangeEvent, FormEvent } from 'react';
import { useNavigate, Link as RouterLink } from 'react-router-dom';
import {
  Box, Button, TextField, Typography, Alert, Paper, Link, CircularProgress, Grid,
} from '@mui/material';
import PersonAddIcon from '@mui/icons-material/PersonAdd';
import { useAppDispatch, useAppSelector } from '../store/hooks';
import { registerThunk, clearError } from '../store/slices/authSlice';
import type { RegisterPayload } from '../api/authApi';

export default function RegisterPage(): React.JSX.Element {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { isAuthenticated, loading, error } = useAppSelector((s) => s.auth);
  const [form, setForm] = useState<RegisterPayload>({
    email: '',
    phone: '',
    password: '',
    first_name: '',
    last_name: '',
  });

  useEffect(() => {
    if (isAuthenticated) navigate('/', { replace: true });
    return () => { dispatch(clearError()); };
  }, [isAuthenticated, navigate, dispatch]);

  const handleChange = (e: ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    dispatch(registerThunk(form));
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
      <Paper elevation={3} sx={{ p: 4, width: 460 }}>
        <Box textAlign="center" mb={3}>
          <PersonAddIcon sx={{ fontSize: 40, color: 'primary.main', mb: 1 }} />
          <Typography variant="h5" fontWeight={700}>Create Account</Typography>
          <Typography variant="body2" color="text.secondary">Payment System</Typography>
        </Box>

        {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

        <Box component="form" onSubmit={handleSubmit}>
          <Grid container spacing={2}>
            <Grid item xs={6}>
              <TextField
                fullWidth label="First Name" name="first_name"
                value={form.first_name} onChange={handleChange} required autoFocus
              />
            </Grid>
            <Grid item xs={6}>
              <TextField
                fullWidth label="Last Name" name="last_name"
                value={form.last_name} onChange={handleChange} required
              />
            </Grid>
          </Grid>
          <TextField
            fullWidth label="Email" name="email" type="email"
            value={form.email} onChange={handleChange} required margin="normal"
            autoComplete="email"
          />
          <TextField
            fullWidth label="Phone" name="phone"
            value={form.phone} onChange={handleChange} required margin="normal"
            placeholder="+1234567890"
          />
          <TextField
            fullWidth label="Password" name="password" type="password"
            value={form.password} onChange={handleChange} required margin="normal"
            autoComplete="new-password"
          />
          <Button
            fullWidth variant="contained" type="submit"
            sx={{ mt: 2, py: 1.2 }} disabled={loading}
          >
            {loading ? <CircularProgress size={22} color="inherit" /> : 'Register'}
          </Button>
        </Box>

        <Typography mt={3} textAlign="center" variant="body2">
          Already have an account?{' '}
          <Link component={RouterLink} to="/login" fontWeight={600}>
            Sign In
          </Link>
        </Typography>
      </Paper>
    </Box>
  );
}
