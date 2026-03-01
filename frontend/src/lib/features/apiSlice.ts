import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { 
  CVEWithSource, 
  CVEDetailResponse,
  PaginatedResponse, 
  StatsResponse, 
  DashboardResponse,
  User,
  UpdateUserRequest,
  Vendor,
  Product,
  SubscriptionTarget,
  ApiError
} from '../types';

export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ 
    baseUrl: process.env.NEXT_PUBLIC_API_URL || '/api/v1',
    prepareHeaders: (headers) => {
      headers.set('Accept', 'application/json');
      return headers;
    },
    credentials: 'include',
  }),
  tagTypes: ['CVE', 'User', 'Vendor', 'Product', 'Stats', 'Subscription'],
  endpoints: (builder) => ({
    getMe: builder.query<User, void>({
      query: () => '/me',
      providesTags: ['User'],
      transformErrorResponse: (response) => response.status === 401 ? { status: 401, data: 'unauthorized' } : response,
    }),
    updateMe: builder.mutation<User, UpdateUserRequest>({
      query: (data) => ({ url: '/profile', method: 'PUT', body: data }),
      invalidatesTags: ['User'],
    }),
    logoutApi: builder.mutation<void, void>({
      query: () => ({ url: '/auth/logout', method: 'POST' }),
      invalidatesTags: ['User', 'Subscription'],
    }),
    getStats: builder.query<StatsResponse, { range?: string }>({
      query: (params) => ({ url: '/stats', params }),
      providesTags: ['Stats'],
    }),
    getCVEs: builder.query<PaginatedResponse<CVEWithSource>, { q?: string; page?: number; start_date?: string; end_date?: string; vendor?: string; product?: string }>({
      query: (params) => ({ url: '/cves', params }),
      providesTags: ['CVE'],
    }),
    getCVEDetail: builder.query<CVEDetailResponse, string>({
      query: (id) => `/cves/${id}`,
      providesTags: (result, error, id) => [{ type: 'CVE', id }],
    }),
    getVendors: builder.query<PaginatedResponse<Vendor>, { q?: string; page?: number; vendor?: string; product?: string }>({
      query: (params) => ({ url: '/vendors', params }),
      providesTags: ['Vendor', 'Subscription'],
    }),
    getProducts: builder.query<PaginatedResponse<Product>, { q?: string; page?: number; vendor?: string; product?: string }>({
      query: (params) => ({ url: '/products', params }),
      providesTags: ['Product', 'Subscription'],
    }),
    getMyDashboard: builder.query<DashboardResponse, void>({
      query: () => '/my/dashboard',
      providesTags: ['Subscription'],
    }),
    subscribe: builder.mutation<{ status: string }, SubscriptionTarget>({
      query: (data) => ({ url: '/subscriptions', method: 'POST', body: data }),
      invalidatesTags: ['Subscription', 'Vendor', 'Product'],
    }),
    unsubscribe: builder.mutation<{ status: string }, string>({
      query: (id) => ({ url: `/subscriptions?id=${id}`, method: 'DELETE' }),
      invalidatesTags: ['Subscription', 'Vendor', 'Product'],
    }),
    updateThreshold: builder.mutation<{ status: string }, { id: string; threshold: string }>({
      query: (data) => ({ url: '/subscriptions/threshold', method: 'PATCH', body: data }),
      invalidatesTags: ['Subscription'],
    }),
    getAdminUsers: builder.query<User[], void>({
      query: () => '/admin/users',
      providesTags: ['User'],
    }),
    updateUserRole: builder.mutation<{ status: string }, { user_id: string; role: string }>({
      query: (data) => ({ url: '/admin/users/role', method: 'PATCH', body: data }),
      invalidatesTags: ['User'],
    }),
  }),
});

export const { 
  useGetMeQuery, 
  useUpdateMeMutation,
  useLogoutApiMutation,
  useGetStatsQuery, 
  useGetCVEsQuery, 
  useGetCVEDetailQuery,
  useGetVendorsQuery,
  useGetProductsQuery,
  useGetMyDashboardQuery,
  useSubscribeMutation,
  useUnsubscribeMutation,
  useUpdateThresholdMutation,
  useGetAdminUsersQuery,
  useUpdateUserRoleMutation,
} = apiSlice;
