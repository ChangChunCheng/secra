import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ 
    baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081/api/v1',
    prepareHeaders: (headers) => {
      // Logic to add token from cookies or state if needed
      return headers;
    },
  }),
  tagTypes: ['CVE', 'User', 'Vendor', 'Product'],
  endpoints: (builder) => ({
    getMe: builder.query<any, void>({
      query: () => '/me',
      providesTags: ['User'],
    }),
    getCVEs: builder.query<any, { q?: string; page?: number }>({
      query: (params) => ({
        url: '/cves',
        params,
      }),
      providesTags: (result) =>
        result
          ? [...result.data.map(({ id }: any) => ({ type: 'CVE' as const, id })), 'CVE']
          : ['CVE'],
    }),
    getVendors: builder.query<any[], void>({
      query: () => '/vendors',
      providesTags: ['Vendor'],
    }),
  }),
});

export const { useGetMeQuery, useGetCVEsQuery, useGetVendorsQuery } = apiSlice;
