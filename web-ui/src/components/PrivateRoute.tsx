import React from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { useAppSelector } from '../store/hooks';

export default function PrivateRoute(): React.JSX.Element {
  const { isAuthenticated } = useAppSelector((s) => s.auth);
  return isAuthenticated ? <Outlet /> : <Navigate to="/login" replace />;
}
