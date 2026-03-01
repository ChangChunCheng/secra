'use client';
import React, { useState } from 'react';
import { useGetCVEsQuery } from '@/lib/features/apiSlice';
import { Search, ShieldAlert, Activity, Users, Box } from 'lucide-react';

export default function Dashboard() {
  const [searchTerm, setSearchTerm] = useState('');
  const { data, isLoading, error } = useGetCVEsQuery({ q: searchTerm, page: 1 });

  return (
    <div className="min-h-screen p-8 font-[family-name:var(--font-geist-mono)]">
      {/* Header */}
      <header className="flex justify-between items-center mb-12 border-b border-green-900 pb-6">
        <div>
          <h1 className="text-4xl font-black tracking-tighter text-green-400">SECRA_SYSTEM_V2</h1>
          <p className="text-xs text-green-800 uppercase tracking-widest">Next-Gen Decoupled Interface</p>
        </div>
        <div className="flex gap-4">
          <div className="px-4 py-2 border border-green-900 bg-green-950/20 rounded text-xs">
            STATUS: <span className="text-green-400 animate-pulse">OPERATIONAL</span>
          </div>
        </div>
      </header>

      {/* Stats Quick Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
        {[
          { label: 'TOTAL_CVES', val: data?.total || '...', icon: ShieldAlert, color: 'text-red-500' },
          { label: 'MONITORED_VENDORS', val: '...', icon: Users, color: 'text-blue-500' },
          { label: 'ACTIVE_PRODUCTS', val: '...', icon: Box, color: 'text-yellow-500' },
        ].map((stat, i) => (
          <div key={i} className="bg-green-950/5 border border-green-900/50 p-6 rounded-lg relative overflow-hidden group hover:border-green-400 transition-colors">
            <stat.icon className={`absolute -right-4 -bottom-4 w-24 h-24 opacity-5 group-hover:opacity-10 transition-opacity`} />
            <div className="text-xs text-green-800 mb-2">{stat.label}</div>
            <div className={`text-3xl font-bold ${stat.color}`}>{stat.val}</div>
          </div>
        ))}
      </div>

      {/* Main Content */}
      <div className="bg-green-950/5 border border-green-900 rounded-lg p-6">
        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
          <h2 className="text-xl font-bold flex items-center gap-2">
            <Activity className="w-5 h-5" /> RECENT_VULNERABILITIES
          </h2>
          
          <div className="relative w-full md:w-96">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-green-800" />
            <input 
              type="text" 
              placeholder="Filter by ID or Description..."
              className="w-full bg-black border border-green-900 rounded py-2 pl-10 pr-4 text-sm focus:outline-none focus:border-green-400 transition-colors"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>

        {/* CVE Table */}
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="text-xs text-green-800 border-b border-green-900">
                <th className="pb-4 font-normal px-2">ID</th>
                <th className="pb-4 font-normal px-2">SEVERITY</th>
                <th className="pb-4 font-normal px-2">DESCRIPTION</th>
                <th className="pb-4 font-normal px-2 text-right">PUBLISHED</th>
              </tr>
            </thead>
            <tbody className="text-sm">
              {isLoading ? (
                <tr><td colSpan={4} className="py-8 text-center animate-pulse">LOADING_DATA...</td></tr>
              ) : data?.data?.map((cve: any) => (
                <tr key={cve.id} className="border-b border-green-900/30 hover:bg-green-400/5 transition-colors group">
                  <td className="py-4 px-2 font-bold text-green-400">{cve.source_uid}</td>
                  <td className="py-4 px-2">
                    <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                      cve.severity === 'CRITICAL' ? 'bg-red-900/50 text-red-400 border border-red-500' :
                      cve.severity === 'HIGH' ? 'bg-orange-900/50 text-orange-400 border border-orange-500' :
                      'bg-green-900/50 text-green-400 border border-green-500'
                    }`}>
                      {cve.severity}
                    </span>
                  </td>
                  <td className="py-4 px-2 max-w-xl truncate text-green-100/70 group-hover:text-green-100 transition-colors">
                    {cve.title || cve.description}
                  </td>
                  <td className="py-4 px-2 text-right text-green-800 font-mono">
                    {new Date(cve.published_at).toISOString().split('T')[0]}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
