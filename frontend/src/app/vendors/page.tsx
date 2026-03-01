'use client';
import { useState, useEffect } from 'react';
import { useGetVendorsQuery, useSubscribeMutation, useUnsubscribeMutation } from '@/lib/features/apiSlice';
import { useSelector } from 'react-redux';
import { RootState } from '@/lib/store';
import { Database, Search, Loader2, Lock, Package, Filter, ChevronDown, ChevronUp } from 'lucide-react';
import Link from 'next/link';
import Pagination from '@/components/Pagination';
import ViewToggle, { ViewMode } from '@/components/ViewToggle';

export default function VendorsPage() {
  const [mounted, setMounted] = useState(false);
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState({ vendor: '', product: '' });
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [showFilters, setShowFilters] = useState(false);
  
  const { data, isLoading, isError, refetch } = useGetVendorsQuery({ ...filters, page });
  const [subscribe, { isLoading: isSubscribing }] = useSubscribeMutation();
  const [unsubscribe, { isLoading: isUnsubscribing }] = useUnsubscribeMutation();
  const { isAuthenticated } = useSelector((state: RootState) => state.auth);

  useEffect(() => { setMounted(true); }, []);

  const handleFilterChange = (key: string, value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setPage(1);
  };

  const handleSubscribe = async (vendorId: string, vendorName: string) => {
    setActionLoading(vendorId);
    try {
      await subscribe({ target_type: 'vendor', target_id: vendorId }).unwrap();
      // Force refetch to update UI immediately
      await refetch();
    } catch (error) {
      console.error('Failed to subscribe:', error);
      alert(`Failed to subscribe to ${vendorName}`);
    } finally {
      setActionLoading(null);
    }
  };

  const handleUnsubscribe = async (subscriptionId: string, vendorName: string) => {
    if (!confirm(`Are you sure you want to unsubscribe from ${vendorName}?`)) return;
    setActionLoading(subscriptionId);
    try {
      await unsubscribe(subscriptionId).unwrap();
      // Force refetch to update UI immediately
      await refetch();
    } catch (error) {
      console.error('Failed to unsubscribe:', error);
      alert(`Failed to unsubscribe from ${vendorName}`);
    } finally {
      setActionLoading(null);
    }
  };

  if (!mounted) return <div className="bg-black min-h-screen" />;

  const vendors = data?.data || [];

  return (
    <div className="p-4 md:p-8 max-w-7xl mx-auto font-mono text-green-500">
      <div className="mb-8 md:mb-12 space-y-6 md:space-y-8">
        <div className="flex flex-row justify-between items-center gap-4 md:gap-8 border-b border-green-900 pb-6 md:pb-10 bg-green-950/5 p-4 md:p-6 rounded-sm">
          <div className="flex items-center gap-3 md:gap-4 flex-1 min-w-0">
            <Database className="w-7 h-7 md:w-10 md:h-10 text-green-500 flex-shrink-0" />
            <div className="min-w-0">
              <h1 className="text-xl md:text-4xl font-black text-green-400 italic tracking-tighter uppercase">Vendors</h1>
              <p className="text-[8px] md:text-[10px] text-green-800 tracking-[0.2em] md:tracking-[0.4em] uppercase font-bold">Entity Directory</p>
            </div>
          </div>
          <ViewToggle mode={viewMode} onModeChange={setViewMode} />
        </div>

        {/* Search Filters */}
        <div className="bg-black border border-green-900 p-4 md:p-6 rounded-sm shadow-2xl space-y-3 md:space-y-0">
          {/* Main Search - Always Visible */}
          <div className="space-y-2">
            <div className="relative group">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-green-900 group-focus-within:text-green-400 transition-colors" />
              <input 
                type="text" placeholder="SEARCH VENDOR NAME..."
                className="w-full bg-black border border-green-900 rounded-sm py-3 pl-10 pr-4 text-xs text-green-400 focus:border-green-400 outline-none transition-all placeholder:text-green-950 uppercase italic font-bold"
                value={filters.vendor}
                onChange={(e) => handleFilterChange('vendor', e.target.value)}
              />
            </div>
            {/* Mobile Filter Hint */}
            <div className="md:hidden flex items-center justify-between text-[10px] text-green-700">
              <span className="italic">1 additional filter available</span>
              <button 
                onClick={() => setShowFilters(!showFilters)}
                className="text-green-500 hover:text-green-400 underline font-bold uppercase"
              >
                {showFilters ? 'Hide Filter' : 'Show Filter'}
              </button>
            </div>
          </div>

          {/* Product Filter - Collapsible on Mobile */}
          <div className={`transition-all ${showFilters ? 'block' : 'hidden md:block'}`}>
            <div className="relative group">
              <Package className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-green-900 group-focus-within:text-green-400 transition-colors" />
              <input 
                type="text" placeholder="FILTER BY PRODUCT..."
                className="w-full bg-black border border-green-900 rounded-sm py-3 pl-10 pr-4 text-xs text-green-100 focus:border-green-400 outline-none transition-all placeholder:text-green-950 uppercase italic font-bold"
                value={filters.product}
                onChange={(e) => handleFilterChange('product', e.target.value)}
              />
            </div>
          </div>
        </div>
      </div>

      {isLoading ? (
        <div className="py-32 flex justify-center animate-pulse text-green-900 uppercase italic font-black tracking-widest text-xl">Loading Registry...</div>
      ) : isError ? (
        <div className="py-20 text-center text-red-500 border border-dashed border-red-900/30 font-black uppercase italic">Uplink Failure: Registry Offline</div>
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
                    <th className="p-6">Vendor Name</th>
                    <th className="p-6 text-center">Products</th>
                    <th className="p-6 text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="text-sm">
                  {vendors.length === 0 ? (
                    <tr><td colSpan={3} className="py-24 text-center text-green-950 italic font-black uppercase tracking-widest">No Vendors Found</td></tr>
                  ) : vendors.map((v: any) => (
                    <tr key={v.id} className="border-b border-green-900/20 hover:bg-green-400/5 transition-all group">
                      <td className="p-6">
                        <div className="font-bold text-green-100 group-hover:text-green-400 uppercase tracking-tighter">{v.name}</div>
                        <div className="text-[9px] text-green-900 font-mono mt-1 uppercase tracking-widest">REF ID: {String(v.id).split('-')[0]}</div>
                      </td>
                      <td className="p-6 text-center font-black text-green-500 tabular-nums">{v.product_count}</td>
                      <td className="p-6 text-right">
                        {isAuthenticated ? (
                          <button 
                            onClick={() => v.subscription_id ? handleUnsubscribe(v.subscription_id, v.name) : handleSubscribe(v.id, v.name)}
                            disabled={actionLoading === v.id || actionLoading === v.subscription_id || isSubscribing || isUnsubscribing}
                            className={`px-4 py-2 text-[9px] font-black border transition-all disabled:opacity-50 disabled:cursor-not-allowed ${
                              v.subscription_id 
                                ? 'border-red-900 text-red-500 hover:bg-red-950/30' 
                                : 'border-green-500 text-green-500 hover:bg-green-500 hover:text-black'
                            }`}
                          >
                            {(actionLoading === v.id || actionLoading === v.subscription_id) ? (
                              <Loader2 className="w-3 h-3 animate-spin inline-block" />
                            ) : (
                              v.subscription_id ? 'UNSUBSCRIBE' : 'SUBSCRIBE'
                            )}
                          </button>
                        ) : (
                          <span className="text-[9px] text-green-900 italic font-black uppercase tracking-tighter">Sign In to Subscribe</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-4">
              {vendors.map((vendor: any) => (
                <div key={vendor.id} className="bg-black border border-green-900/50 p-6 rounded-sm hover:border-green-400 transition-all group relative overflow-hidden shadow-xl">
                  <div className="flex justify-between items-start mb-10">
                    <div>
                      <h3 className="font-bold text-lg text-green-100 group-hover:text-green-400 transition-colors uppercase truncate max-w-[180px]">{vendor.name}</h3>
                      <div className="text-[9px] text-green-900 font-mono mt-1 uppercase tracking-tighter">REF: {String(vendor?.id || '').split('-')[0]}</div>
                    </div>
                    <span className="text-[9px] bg-green-900/20 text-green-500 px-2 py-1 border border-green-800 font-black uppercase tracking-widest">
                      {vendor.product_count || 0} Assets
                    </span>
                  </div>
                  {isAuthenticated ? (
                    <button 
                      onClick={() => vendor.subscription_id ? handleUnsubscribe(vendor.subscription_id, vendor.name) : handleSubscribe(vendor.id, vendor.name)}
                      disabled={actionLoading === vendor.id || actionLoading === vendor.subscription_id || isSubscribing || isUnsubscribing}
                      className={`w-full py-3 text-[10px] font-black border transition-all disabled:opacity-50 disabled:cursor-not-allowed ${
                        vendor.subscription_id 
                          ? 'border-red-900 text-red-500 hover:bg-red-950/30' 
                          : 'border-green-500 text-green-500 hover:bg-green-500 hover:text-black shadow-[0_0_15px_rgba(34,197,94,0.1)]'
                      }`}
                    >
                      {(actionLoading === vendor.id || actionLoading === vendor.subscription_id) ? (
                        <><Loader2 className="w-3 h-3 animate-spin inline-block mr-2" />PROCESSING...</>
                      ) : (
                        vendor.subscription_id ? 'UNSUBSCRIBE' : 'SUBSCRIBE VENDOR'
                      )}
                    </button>
                  ) : (
                    <Link href="/login" className="block w-full py-3 text-center border border-dashed border-green-900/30 text-[9px] text-green-900 uppercase italic font-black hover:text-green-400 transition-colors">
                      Sign In to Subscribe
                    </Link>
                  )}
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
    </div>
  );
}
