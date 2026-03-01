'use client';
import { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { useGetMeQuery } from '@/lib/features/apiSlice';
import { setCredentials, logout } from '@/lib/features/authSlice';

export default function AuthInit({ children }: { children: React.ReactNode }) {
  const dispatch = useDispatch();
  // CRITICAL: Fetch profile on every mount/refresh to ensure role synchronization
  const { data: user, isSuccess, isError, error } = useGetMeQuery();

  useEffect(() => {
    if (isSuccess && user) {
      // Force update state with latest data from DB (handles role changes)
      dispatch(setCredentials(user));
    } else if (isError && (error as any)?.status === 401) {
      // Clear local state if backend session is expired/invalid
      dispatch(logout());
    }
  }, [isSuccess, user, isError, error, dispatch]);

  return <>{children}</>;
}
