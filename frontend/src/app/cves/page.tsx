'use client';
import React, { useState, useEffect } from 'react';
import { useGetCVEsQuery, useGetCVEDetailQuery } from '@/lib/features/apiSlice';
import { Search, X, ExternalLink, Shield, Calendar, Database, Package, Filter, ChevronDown, ChevronUp } from 'lucide-react';
import Pagination from '@/components/Pagination';
import ViewToggle, { ViewMode } from '@/components/ViewToggle';

// --- Safe Rendering Utilities ---
const formatSafeDate = (d: any) => {
  if (!d) return 'N/A';
  try {
    const date = new Date(d);
    return isNaN(date.getTime()) ? 'N/A' : date.toISOString().split('T')[0];
  } catch (e) { return 'N/A'; }
};

const formatSafeScore = (s: any) => {
  if (s === null || s === undefined) return '—';
  const num = parseFloat(s);
  return isNaN(num) ? '—' : num.toFixed(1);
};

// --- Shared Components ---
function StatusBadge({ severity }: { severity: string }) {
  const colors: Record<string, string> = {
    CRITICAL: 'bg-red-900/20 text-red-500 border-red-500',
    HIGH: 'bg-orange-900/20 text-orange-500 border-orange-500',
    MEDIUM: 'bg-yellow-900/20 text-yellow-500 border-yellow-500',
    LOW: 'bg-green-900/20 text-green-500 border-green-500',
  };
  const label = (severity || 'UNKNOWN').toUpperCase();
  return (
    <span className={`px-2 py-0.5 rounded-sm text-[9px] font-black border uppercase tracking-tighter ${colors[label] || 'bg-green-900/20 text-green-500 border-green-500'}`}>
      {label}
    </span>
  );
}

function CVSSScoreDisplay({ score }: { score: any }) {
  const val = parseFloat(score);
  if (isNaN(val)) return <span className="text-green-900">—</span>;
  const color = val >= 9.0 ? 'text-red-500' : val >= 7.0 ? 'text-orange-500' : val >= 4.0 ? 'text-yellow-500' : 'text-green-500';
  return <span className={`font-black tabular-nums ${color}`}>{val.toFixed(1)}</span>;
}

// --- Shared Modal ---
function CVEDetailModal({ id, onClose }: { id: string, onClose: () => void }) {
  const { data, isLoading, isError } = useGetCVEDetailQuery(id);
  if (!id) return null;
  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/95 backdrop-blur-md">
      <div className="w-full max-w-4xl bg-black border border-green-500 rounded-sm overflow-hidden flex flex-col max-h-[90vh] shadow-2xl font-mono">
        <div className="p-6 border-b border-green-900 bg-green-950/20 flex justify-between items-center text-green-500">
          <div className="flex items-center gap-4">
            <h2 className="text-2xl font-black italic tracking-tighter uppercase">{data?.cve?.source_uid || 'Loading...'}</h2>
            {data?.cve?.severity && <StatusBadge severity={data.cve.severity} />}
            <div className="flex items-center gap-2 bg-green-900/10 px-3 py-1 border border-green-900/30 rounded-sm">
              <span className="text-[9px] font-black text-green-800 uppercase">CVSS</span>
              <CVSSScoreDisplay score={data?.cve?.cvss_score} />
            </div>
          </div>
          <button onClick={onClose} className="text-green-800 hover:text-white transition-colors"><X className="w-6 h-6" /></button>
        </div>
        <div className="flex-1 overflow-y-auto p-8 space-y-8 text-green-100 font-sans custom-scrollbar">
          {isLoading ? <div className="py-20 text-center animate-pulse text-green-800 font-black uppercase tracking-widest font-mono">Retrieving Data...</div> : (
            <div className="font-mono">
              <section className="mb-8">
                <h3 className="text-[10px] text-green-800 font-black uppercase mb-3 tracking-widest italic underline">Summary</h3>
                <p className="text-sm leading-relaxed opacity-90 font-sans">{data?.cve?.description || 'No description provided.'}</p>
              </section>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
                <section>
                  <h3 className="text-[10px] text-green-800 font-black uppercase mb-3 italic">Affected Products</h3>
                  <div className="space-y-2">
                    {data?.products && data.products.length > 0 ? data.products.map((p: any, i: number) => (
                      <div key={i} className="flex justify-between items-center bg-green-900/5 border border-green-900/30 p-2 rounded-sm text-[10px] hover:border-green-400 group">
                        <span className="text-green-100 uppercase group-hover:text-white">{p.name || 'Unknown'}</span>
                        <span className="text-yellow-600 font-black whitespace-nowrap uppercase">{p.vendor_name || 'Unknown'}</span>
                      </div>
                    )) : <div className="text-[10px] text-green-900 italic uppercase">None linked</div>}
                  </div>
                </section>
                <section>
                  <h3 className="text-[10px] text-green-800 font-black uppercase mb-3 italic">Vulnerability Info</h3>
                  <div className="flex flex-wrap gap-2">
                    {data?.weaknesses && data.weaknesses.length > 0 ? data.weaknesses.map((w: any, i: number) => (
                      <span key={i} className="px-2 py-1 bg-blue-900/20 border border-blue-800 text-blue-400 text-[9px] font-black uppercase">{w.weakness_type}</span>
                    )) : <div className="text-[10px] text-green-900 italic uppercase">Unclassified</div>}
                  </div>
                </section>
              </div>
              <section>
                <h3 className="text-[10px] text-green-800 font-black uppercase mb-3 italic">References</h3>
                <div className="grid grid-cols-1 gap-2">
                  {data?.references?.map((r: any, i: number) => (
                    <a key={i} href={r.url} target="_blank" rel="noreferrer" className="text-[10px] text-green-100/50 hover:text-green-400 flex items-center gap-2 truncate transition-colors">
                      <ExternalLink className="w-3 h-3 flex-shrink-0" /> {r.url}
                    </a>
                  ))}
                </div>
              </section>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default function CVEPage() {
  const [mounted, setMounted] = useState(false);
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [selectedCVE, setSelectedCVE] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState({ q: '', start_date: '', end_date: '', vendor: '', product: '' });
  const [showFilters, setShowFilters] = useState(false);

  const { data, isLoading, isError } = useGetCVEsQuery({ ...filters, page });

  useEffect(() => { setMounted(true); }, []);

  const handleFilterChange = (key: string, value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setPage(1);
  };

  if (!mounted) return <div className="bg-black min-h-screen" />;

  const cves = Array.isArray(data?.data) ? data.data : [];

  return (
    <div className="p-4 md:p-8 max-w-7xl mx-auto font-mono text-green-500">
      <div className="mb-8 md:mb-12 space-y-6 md:space-y-8">
        <div className="flex flex-row justify-between items-center gap-4 md:gap-8 border-b border-green-900 pb-6 md:pb-10 bg-green-950/5 p-4 md:p-6 rounded-sm relative overflow-hidden">
          <div className="flex items-center gap-3 md:gap-4 flex-1 min-w-0">
            <Shield className="w-7 h-7 md:w-10 md:h-10 text-green-500 flex-shrink-0" />
            <div className="min-w-0">
              <h1 className="text-xl md:text-4xl font-black text-green-400 italic tracking-tighter uppercase">CVEs</h1>
              <p className="text-[8px] md:text-[10px] text-green-800 tracking-[0.2em] md:tracking-[0.4em] uppercase font-bold">Security Alerts</p>
            </div>
          </div>
          <ViewToggle mode={viewMode} onModeChange={setViewMode} />
        </div>

        {/* Filter Bar */}
        <div className="bg-black border border-green-900 p-4 md:p-6 rounded-sm shadow-2xl space-y-3 md:space-y-0">
          {/* Main Search - Always Visible */}
          <div className="space-y-2">
            <div className="relative group">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-green-900 group-focus-within:text-green-400 transition-colors" />
              <input 
                type="text" placeholder="SEARCH CVE ID OR CONTENT..."
                className="w-full bg-black border border-green-900 rounded-sm py-2.5 pl-10 pr-4 text-xs text-green-400 focus:border-green-400 outline-none transition-all placeholder:text-green-950 uppercase italic font-bold"
                value={filters.q}
                onChange={(e) => handleFilterChange('q', e.target.value)}
              />
            </div>
            {/* Mobile Filter Hint */}
            <div className="md:hidden flex items-center justify-between text-[10px] text-green-700">
              <span className="italic">4 additional filters available</span>
              <button 
                onClick={() => setShowFilters(!showFilters)}
                className="text-green-500 hover:text-green-400 underline font-bold uppercase"
              >
                {showFilters ? 'Hide Filters' : 'Show Filters'}
              </button>
            </div>
          </div>

          {/* Advanced Filters - Collapsible on Mobile */}
          <div className={`grid grid-cols-1 md:grid-cols-4 gap-3 md:gap-4 transition-all ${showFilters ? 'grid' : 'hidden md:grid'}`}>
            <div className="relative group">
              <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-green-900 group-focus-within:text-green-400 transition-colors" />
              <input 
                type="date"
                lang="en"
                placeholder="From Date"
                className="w-full bg-black border border-green-900 rounded-sm py-2.5 pl-10 pr-4 text-xs text-green-400 focus:border-green-400 outline-none transition-all appearance-none [&::-webkit-calendar-picker-indicator]:invert [&::-webkit-calendar-picker-indicator]:opacity-50"
                value={filters.start_date}
                onChange={(e) => handleFilterChange('start_date', e.target.value)}
              />
            </div>
            <div className="relative group">
              <Calendar className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-green-900 group-focus-within:text-green-400 transition-colors" />
              <input 
                type="date"
                lang="en"
                placeholder="To Date"
                className="w-full bg-black border border-green-900 rounded-sm py-2.5 pl-10 pr-4 text-xs text-green-400 focus:border-green-400 outline-none transition-all appearance-none [&::-webkit-calendar-picker-indicator]:invert [&::-webkit-calendar-picker-indicator]:opacity-50"
                value={filters.end_date}
                onChange={(e) => handleFilterChange('end_date', e.target.value)}
              />
            </div>
            <div className="relative group">
              <Database className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-green-900 group-focus-within:text-green-400 transition-colors" />
              <input 
                type="text" placeholder="FILTER VENDOR..."
                className="w-full bg-black border border-green-900 rounded-sm py-2.5 pl-10 pr-4 text-xs text-green-100 focus:border-green-400 outline-none transition-all placeholder:text-green-950 uppercase font-bold"
                value={filters.vendor}
                onChange={(e) => handleFilterChange('vendor', e.target.value)}
              />
            </div>
            <div className="relative group">
              <Package className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-green-900 group-focus-within:text-green-400 transition-colors" />
              <input 
                type="text" placeholder="FILTER PRODUCT..."
                className="w-full bg-black border border-green-900 rounded-sm py-2.5 pl-10 pr-4 text-xs text-green-100 focus:border-green-400 outline-none transition-all placeholder:text-green-950 uppercase font-bold"
                value={filters.product}
                onChange={(e) => handleFilterChange('product', e.target.value)}
              />
            </div>
          </div>
        </div>
      </div>

      {isLoading ? (
        <div className="py-32 flex justify-center animate-pulse text-green-900 uppercase italic font-black tracking-widest text-xl">Loading records...</div>
      ) : isError ? (
        <div className="py-32 text-center text-red-500 font-black uppercase italic border border-dashed border-red-900/30 font-mono shadow-2xl">Uplink Error: Protocol Mismatch</div>
      ) : (
        <>
          <div className="mb-4 bg-black border border-green-900 rounded-sm shadow-xl overflow-hidden">
            <Pagination 
              currentPage={page}
              totalPages={data?.total_pages || 1}
              totalItems={data?.total || 0}
              onPageChange={setPage}
            />
          </div>

          {viewMode === 'list' ? (
            <div className="bg-black border border-green-900 rounded-sm overflow-hidden shadow-2xl mb-4">
              <table className="w-full text-left">
                <thead>
                  <tr className="text-[10px] text-green-800 uppercase border-b border-green-900/50 bg-green-900/10 font-black tracking-widest">
                    <th className="p-3 md:p-6">CVE ID</th>
                    <th className="p-3 md:p-6 text-center">Severity</th>
                    <th className="p-3 md:p-6 text-center">CVSS</th>
                    <th className="p-3 md:p-6 hidden sm:table-cell">Affected Assets</th>
                    <th className="p-3 md:p-6 hidden lg:table-cell">Summary</th>
                    <th className="p-3 md:p-6 text-right">Published</th>
                  </tr>
                </thead>
                <tbody className="text-xs">
                  {cves.length === 0 ? (
                    <tr><td colSpan={6} className="py-12 md:py-24 text-center text-green-950 italic font-black uppercase tracking-widest text-xs md:text-sm">Null Set: Zero Matches Found</td></tr>
                  ) : cves.map((cve: any) => (
                    <tr key={cve.id || Math.random()} onClick={() => setSelectedCVE(cve.id)} className="border-b border-green-900/20 hover:bg-green-400/5 cursor-pointer group transition-all">
                      <td className="p-3 md:p-6 font-bold text-green-400 group-hover:text-green-100 uppercase text-[10px] md:text-xs">{cve.source_uid || 'N/A'}</td>
                      <td className="p-3 md:p-6 text-center"><StatusBadge severity={cve.severity} /></td>
                      <td className="p-3 md:p-6 text-center"><CVSSScoreDisplay score={cve.cvss_score} /></td>
                      <td className="p-3 md:p-6 hidden sm:table-cell">
                        <div className="flex flex-wrap gap-1 max-w-[200px]">
                          {cve.assets ? cve.assets.split(', ').slice(0, 2).map((a: string, i: number) => (
                            <span key={i} className="text-[8px] bg-green-900/10 text-green-700 px-1 border border-green-900/20 uppercase font-bold font-sans">{a}</span>
                          )) : <span className="text-[8px] text-green-950 italic uppercase">None</span>}
                        </div>
                      </td>
                      <td className="p-3 md:p-6 max-w-md truncate opacity-70 group-hover:opacity-100 font-sans italic hidden lg:table-cell">{cve.title || cve.description || 'No data'}</td>
                      <td className="p-3 md:p-6 text-right text-green-900 font-mono text-[10px] uppercase">
                        {formatSafeDate(cve.published_at)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6 mb-4">
              {cves.map((cve: any) => (
                <div key={cve.id || Math.random()} onClick={() => setSelectedCVE(cve.id)} className="bg-black border border-green-900/50 p-6 rounded-sm hover:border-green-400 transition-all cursor-pointer group shadow-xl relative overflow-hidden">
                  <div className="flex justify-between items-start mb-6">
                    <div>
                      <h3 className="text-lg font-black text-green-400 italic group-hover:text-green-100 transition-colors uppercase tracking-tighter">{cve.source_uid || 'N/A'}</h3>
                      <div className="text-[10px] text-green-900 font-mono mt-1 uppercase tracking-widest">ID_{String(cve.id || '').split('-')[0] || 'N/A'}</div>
                    </div>
                    <div className="flex flex-col items-end gap-2">
                      <StatusBadge severity={cve.severity} />
                      <div className="text-[10px] font-black"><span className="text-green-900 mr-1">CVSS:</span> <CVSSScoreDisplay score={cve.cvss_score} /></div>
                    </div>
                  </div>
                  <p className="text-xs text-green-100 opacity-60 line-clamp-3 mb-8 italic font-sans">{cve.title || cve.description || 'No summary available.'}</p>
                  <div className="flex justify-between items-end">
                    <div className="flex flex-wrap gap-1 max-w-[150px]">
                      {String(cve.assets || '').split(', ').slice(0, 2).filter(Boolean).map((a: string, i: number) => (
                        <span key={i} className="text-[8px] bg-green-900/10 text-green-700 px-1 border border-green-900/20 uppercase font-bold font-sans">{a}</span>
                      ))}
                    </div>
                    <div className="text-[10px] text-green-900 font-mono text-right">{formatSafeDate(cve.published_at)}</div>
                  </div>
                </div>
              ))}
            </div>
          )}

          <div className="bg-black border border-green-900 rounded-sm shadow-xl">
            <Pagination 
              currentPage={page}
              totalPages={data?.total_pages || 1}
              totalItems={data?.total || 0}
              onPageChange={setPage}
            />
          </div>
        </>
      )}

      {selectedCVE && <CVEDetailModal id={selectedCVE} onClose={() => setSelectedCVE(null)} />}
    </div>
  );
}
