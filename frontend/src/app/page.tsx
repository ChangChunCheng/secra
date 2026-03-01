'use client';
import React, { useState, useEffect, useMemo } from 'react';
import { useGetCVEsQuery, useGetStatsQuery, useGetCVEDetailQuery } from '@/lib/features/apiSlice';
import { Search, ShieldAlert, Activity, Shield, Loader2, X, ExternalLink, AlertTriangle } from 'lucide-react';
import { XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area } from 'recharts';

// --- ROBUST UTILITIES ---
const safeFormatDate = (dateStr: any) => {
  if (!dateStr) return 'N/A';
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return 'N/A';
    return d.toISOString().split('T')[0];
  } catch (e) { return 'N/A'; }
};

const safeFormatFullDate = (dateStr: any) => {
  if (!dateStr) return 'N/A';
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return 'N/A';
    return d.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
  } catch (e) { return 'N/A'; }
};

const safeFormatShortDate = (dateStr: any) => {
  if (!dateStr) return '';
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return '';
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  } catch (e) { return ''; }
};

// --- SHARED COMPONENTS ---
function StatusBadge({ severity }: { severity: string }) {
  const colors: Record<string, string> = {
    CRITICAL: 'bg-red-900/20 text-red-500 border-red-500',
    HIGH: 'bg-orange-900/20 text-orange-500 border-orange-500',
    MEDIUM: 'bg-yellow-900/20 text-yellow-500 border-yellow-500',
    LOW: 'bg-green-900/20 text-green-500 border-green-500',
  };
  return (
    <span className={`px-2 py-0.5 rounded-sm text-[9px] font-black border uppercase tracking-tighter ${colors[severity] || 'bg-green-900/20 text-green-500 border-green-500'}`}>
      {severity || 'UNKNOWN'}
    </span>
  );
}

function CVSSScore({ score }: { score: number | null | undefined }) {
  if (score === null || score === undefined) return <span className="text-green-900">—</span>;
  const color = score >= 9.0 ? 'text-red-500' : score >= 7.0 ? 'text-orange-500' : score >= 4.0 ? 'text-yellow-500' : 'text-green-500';
  return <span className={`font-black tabular-nums ${color}`}>{score.toFixed(1)}</span>;
}

// --- CVE DETAIL MODAL ---
function CVEDetailModal({ id, onClose }: { id: string, onClose: () => void }) {
  const { data, isLoading, isError } = useGetCVEDetailQuery(id);
  if (isError) return (
    <div className="fixed inset-0 z-[110] flex items-center justify-center p-4 bg-black/90 font-mono text-green-500">
      <div className="bg-black border border-red-500 p-8 text-center text-red-500 shadow-2xl">
        <AlertTriangle className="mx-auto mb-4" />
        <p className="font-black uppercase">Loading error: Cannot retrieve CVE details</p>
        <button onClick={onClose} className="mt-4 text-xs underline uppercase hover:text-white">Close</button>
      </div>
    </div>
  );

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/95 backdrop-blur-md">
      <div className="w-full max-w-4xl bg-black border border-green-500 rounded-sm overflow-hidden flex flex-col max-h-[90vh] shadow-2xl">
        <div className="p-6 border-b border-green-900 bg-green-950/20 flex justify-between items-center font-mono text-green-500">
          <div className="flex items-center gap-4">
            <h2 className="text-2xl font-black italic tracking-tighter uppercase">{data?.cve?.source_uid || 'Loading...'}</h2>
            {data?.cve?.severity && <StatusBadge severity={data.cve.severity} />}
            <div className="flex items-center gap-2 bg-green-900/10 px-3 py-1 border border-green-900/30 rounded-sm">
              <span className="text-[9px] font-black text-green-800 uppercase">CVSS</span>
              <CVSSScore score={data?.cve?.cvss_score} />
            </div>
          </div>
          <button onClick={onClose} className="text-green-800 hover:text-white transition-colors"><X className="w-6 h-6" /></button>
        </div>
        
        <div className="flex-1 overflow-y-auto p-8 space-y-8 custom-scrollbar font-mono">
          {isLoading ? (
            <div className="py-20 text-center animate-pulse text-green-800 uppercase font-black tracking-widest">Retrieving Data...</div>
          ) : (
            <>
              <section>
                <h3 className="text-[10px] text-green-800 font-black uppercase mb-3 tracking-widest italic underline">CVE Summary</h3>
                <p className="text-sm leading-relaxed text-green-100/90 font-sans">{data?.cve?.description || 'No description available.'}</p>
              </section>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-8 text-green-100">
                <section>
                  <h3 className="text-[10px] text-green-800 font-black uppercase mb-3 italic">Affected Products</h3>
                  <div className="space-y-2">
                    {data?.products && data.products.length > 0 ? data.products.map((p: any, i: number) => (
                      <div key={i} className="flex justify-between items-center bg-green-900/5 border border-green-900/30 p-2 rounded-sm text-[10px] hover:border-green-400 group">
                        <span className="uppercase group-hover:text-white">{p.name}</span>
                        <span className="text-yellow-600 font-black uppercase whitespace-nowrap">{p.vendor_name}</span>
                      </div>
                    )) : <p className="text-[10px] text-green-900 italic uppercase">None identified</p>}
                  </div>
                </section>
                <section>
                  <h3 className="text-[10px] text-green-800 font-black uppercase mb-3 italic">Vulnerability Types</h3>
                  <div className="flex flex-wrap gap-2">
                    {data?.weaknesses && data.weaknesses.length > 0 ? data.weaknesses.map((w: any, i: number) => (
                      <span key={i} className="px-2 py-1 bg-blue-900/20 border border-blue-800 text-blue-400 text-[9px] font-black uppercase">{w.weakness_type}</span>
                    )) : <p className="text-[10px] text-green-900 italic uppercase">Unclassified</p>}
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
            </>
          )}
        </div>
      </div>
    </div>
  );
}

// --- MAIN DASHBOARD ---
export default function Dashboard() {
  const [mounted, setMounted] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [timeRange, setTimeRange] = useState('1m');
  const [selectedCVE, setSelectedCVE] = useState<string | null>(null);

  const { data: cveData, isLoading: cvesLoading, isError: cveError } = useGetCVEsQuery({ q: searchTerm, page: 1 });
  const { data: stats, isLoading: statsLoading } = useGetStatsQuery({ range: timeRange });

  useEffect(() => { setMounted(true); }, []);

  const chartData = useMemo(() => {
    if (!stats?.chart_data) return [];
    return stats.chart_data.map((d: any) => ({
      period: safeFormatShortDate(d.period),
      count: d.count,
      label: safeFormatFullDate(d.period)
    }));
  }, [stats]);

  if (!mounted) return <div className="bg-black min-h-screen" />;

  return (
    <div className="p-8 max-w-7xl mx-auto font-mono text-green-500">
      <header className="flex flex-col md:flex-row justify-between items-start md:items-end mb-12 border-b border-green-900/50 pb-8 gap-6">
        <div>
          <div className="flex items-center gap-3">
            <Shield className="w-10 h-10 text-green-500" />
            <h1 className="text-5xl font-black tracking-tighter text-green-400 italic uppercase">SECRA</h1>
          </div>
          <p className="text-[9px] text-green-800 tracking-[0.6em] uppercase pl-1 font-bold mt-2">Security Risk & Asset Monitor</p>
        </div>
        <div className="text-right hidden md:block">
          <div className="text-[10px] text-green-800 uppercase font-black mb-1">Status</div>
          <div className="text-xs text-green-400 border border-green-900 px-4 py-1 rounded-full bg-green-950/20 uppercase">● Online</div>
        </div>
      </header>

      {/* Global Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        {[
          { label: 'Total CVEs', val: stats?.total_cves || 0, color: 'text-red-500' },
          { label: 'Vendors', val: stats?.total_vendors || 0, color: 'text-blue-500' },
          { label: 'Products', val: stats?.total_products || 0, color: 'text-yellow-500' },
        ].map((stat, i) => (
          <div key={i} className="bg-black border border-green-900/40 p-6 rounded-sm shadow-xl hover:border-green-400 transition-all group">
            <div className="text-[10px] text-green-900 font-black mb-2 tracking-widest uppercase">{stat.label}</div>
            <div className={`text-4xl font-black ${stat.color} tracking-tighter tabular-nums`}>
              {statsLoading ? '...' : stat.val.toLocaleString()}
            </div>
          </div>
        ))}
      </div>

      {/* Threat Chart */}
      <div className="bg-black border border-green-900 rounded-sm p-8 mb-12 shadow-2xl">
        <div className="flex justify-between items-center mb-10">
          <h3 className="text-xs font-black text-green-800 uppercase flex items-center gap-3 italic tracking-widest">
            <Activity className="w-4 h-4 text-green-500" /> Vulnerability Trends
          </h3>
          <div className="flex border border-green-900 p-1 rounded-sm bg-black">
            {['1w', '1m', '1y', '5y'].map((r) => (
              <button key={r} onClick={() => setTimeRange(r)} className={`px-3 py-1 text-[9px] font-black transition-all ${timeRange === r ? 'bg-green-500 text-black' : 'text-green-900 hover:text-green-400'}`}>
                {r.toUpperCase()}
              </button>
            ))}
          </div>
        </div>
        <div className="h-[350px] w-full font-sans text-xs">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData}>
              <defs>
                <linearGradient id="neon" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#22c55e" stopOpacity={0.2}/><stop offset="95%" stopColor="#22c55e" stopOpacity={0}/>
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#052e16" vertical={false} />
              <XAxis 
                dataKey="period" 
                stroke="#064e3b" 
                fontSize={9} 
                tickLine={false} 
                axisLine={false}
                angle={-45}
                textAnchor="end"
                height={60}
                interval="preserveStartEnd"
              />
              <YAxis stroke="#064e3b" fontSize={10} tickLine={false} axisLine={false} />
              <Tooltip 
                content={({ active, payload }: any) => {
                  if (active && payload && payload.length) {
                    return (
                      <div className="bg-black border border-green-500 p-3 shadow-2xl font-mono">
                        <p className="text-[10px] text-green-800 font-black mb-1 uppercase tracking-tighter">{payload[0].payload.label}</p>
                        <p className="text-sm font-bold text-green-400 uppercase tracking-tighter">Detected: <span className="text-white">{payload[0].value}</span></p>
                      </div>
                    );
                  }
                  return null;
                }}
              />
              <Area type="monotone" dataKey="count" stroke="#4ade80" fill="url(#neon)" strokeWidth={3} dot={{ r: 3, fill: '#000', stroke: '#4ade80' }} />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Main Alerts Feed */}
      <div className="bg-black border border-green-900 rounded-sm shadow-2xl overflow-hidden">
        <div className="p-6 border-b border-green-900 flex flex-col md:flex-row justify-between items-center bg-green-950/10 gap-4">
          <h2 className="text-sm font-black text-green-400 italic uppercase">Recent CVE Alerts</h2>
          <div className="relative w-full md:w-96 group">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-green-900 group-focus-within:text-green-400" />
            <input 
              type="text" placeholder="Search by ID or product..."
              className="bg-black border border-green-900 rounded-sm py-3 pl-10 pr-4 text-xs text-green-400 focus:border-green-400 w-full outline-none transition-all placeholder:text-green-950 uppercase italic"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="text-[10px] text-green-800 uppercase border-b border-green-900/50 bg-green-900/5 font-black tracking-widest">
                <th className="p-6">CVE ID</th>
                <th className="p-6 text-center">Severity</th>
                <th className="p-6 text-center">CVSS</th>
                <th className="p-6">Products</th>
                <th className="p-6">Description</th>
                <th className="p-6 text-right">Published</th>
              </tr>
            </thead>
            <tbody className="text-xs">
              {cvesLoading ? (
                <tr><td colSpan={6} className="py-24 text-center text-green-900 animate-pulse font-black uppercase tracking-widest">Loading...</td></tr>
              ) : cveError ? (
                <tr><td colSpan={6} className="py-24 text-center text-red-500 font-black italic uppercase">Load Error</td></tr>
              ) : cveData?.data?.map((cve: any) => (
                <tr key={cve.id} onClick={() => setSelectedCVE(cve.id)} className="border-b border-green-900/20 hover:bg-green-400/5 transition-all cursor-pointer group">
                  <td className="p-6 font-bold text-green-400 font-mono tracking-tighter group-hover:text-green-100 uppercase">{cve.source_uid}</td>
                  <td className="p-6 text-center"><StatusBadge severity={cve.severity} /></td>
                  <td className="p-6 text-center"><CVSSScore score={cve.cvss_score} /></td>
                  <td className="p-6 max-w-[180px]">
                    <div className="flex flex-wrap gap-1">
                      {cve.assets ? cve.assets.split(', ').slice(0, 2).map((a: string, i: number) => (
                        <span key={i} className="text-[8px] bg-green-900/10 text-green-700 px-1 border border-green-900/20 uppercase font-bold font-sans">{a}</span>
                      )) : <span className="text-[8px] text-green-950 italic uppercase">None</span>}
                    </div>
                  </td>
                  <td className="p-6 max-w-md truncate text-green-100/60 group-hover:text-green-100 font-sans italic opacity-80">{cve.title || cve.description}</td>
                  <td className="p-6 text-right text-green-900 font-mono text-[10px] uppercase">{safeFormatDate(cve.published_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
      {selectedCVE && <CVEDetailModal id={selectedCVE} onClose={() => setSelectedCVE(null)} />}
    </div>
  );
}
