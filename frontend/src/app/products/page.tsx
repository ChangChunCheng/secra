'use client';
import { useState, useEffect } from 'react';
import { useGetProductsQuery, useSubscribeMutation, useUnsubscribeMutation } from '@/lib/features/apiSlice';
import { useSelector } from 'react-redux';
import { RootState } from '@/lib/store';
import { Search, Loader2, Package, Lock, Database, Filter, ChevronDown, ChevronUp } from 'lucide-react';
import Link from 'next/link';
import Pagination from '@/components/Pagination';
import ViewToggle, { ViewMode } from '@/components/ViewToggle';

export default function ProductsPage() {
  const [mounted, setMounted] = useState(false);
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState({ product: '', vendor: '' });
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [showFilters, setShowFilters] = useState(false);
  
  const { data, isLoading, isError, refetch } = useGetProductsQuery({ ...filters, page });
  const [subscribe, { isLoading: isSubscribing }] = useSubscribeMutation();
  const [unsubscribe, { isLoading: isUnsubscribing }] = useUnsubscribeMutation();
  const { isAuthenticated } = useSelector((state: RootState) => state.auth);

  useEffect(() => { setMounted(true); }, []);

  const handleFilterChange = (key: string, value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setPage(1);
  };

  const handleSubscribe = async (productId: string, productName: string) => {
    setActionLoading(productId);
    try {
      await subscribe({ target_type: 'product', target_id: productId }).unwrap();
      // Force refetch to update UI immediately
      await refetch();
    } catch (error) {
      console.error('Failed to subscribe:', error);
      alert(`Failed to subscribe to ${productName}`);
    } finally {
      setActionLoading(null);
    }
  };

  const handleUnsubscribe = async (subscriptionId: string, productName: string) => {
    if (!confirm(`Are you sure you want to unsubscribe from ${productName}?`)) return;
    setActionLoading(subscriptionId);
    try {
      await unsubscribe(subscriptionId).unwrap();
      // Force refetch to update UI immediately
      await refetch();
    } catch (error) {
      console.error('Failed to unsubscribe:', error);
      alert(`Failed to unsubscribe from ${productName}`);
    } finally {
      setActionLoading(null);
    }
  };

  if (!mounted) return <div className="bg-black min-h-screen" />;

  const products = data?.data || [];

  return (
    <div className="p-4 md:p-8 max-w-7xl mx-auto font-mono text-green-500">
      <div className="mb-8 md:mb-12 space-y-6 md:space-y-8">
        <div className="flex flex-row justify-between items-center gap-4 md:gap-8 border-b border-green-900 pb-6 md:pb-10 bg-green-950/5 p-4 md:p-6 rounded-sm">
          <div className="flex items-center gap-3 md:gap-4 flex-1 min-w-0">
            <Package className="w-7 h-7 md:w-10 md:h-10 text-yellow-500 flex-shrink-0" />
            <div className="min-w-0">
              <h1 className="text-xl md:text-4xl font-black text-yellow-500 uppercase italic tracking-tighter">Products</h1>
              <p className="text-[8px] md:text-[10px] text-green-800 tracking-[0.2em] md:tracking-[0.4em] uppercase font-bold">Asset Inventory</p>
            </div>
          </div>
          <ViewToggle mode={viewMode} onModeChange={setViewMode} />
        </div>

        {/* Search Filters */}
        <div className="bg-black border border-green-900 p-4 md:p-6 rounded-sm shadow-2xl space-y-3 md:space-y-0">
          {/* Main Search - Always Visible */}
          <div className="space-y-2">
            <div className="relative group">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-green-900 group-focus-within:text-yellow-500 transition-colors" />
              <input 
                type="text" placeholder="SEARCH PRODUCT NAME..."
                className="w-full bg-black border border-green-900 rounded-sm py-3 pl-10 pr-4 text-xs text-green-400 focus:border-yellow-500 outline-none transition-all placeholder:text-green-950 uppercase italic font-bold"
                value={filters.product}
                onChange={(e) => handleFilterChange('product', e.target.value)}
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

          {/* Vendor Filter - Collapsible on Mobile */}
          <div className={`transition-all ${showFilters ? 'block' : 'hidden md:block'}`}>
            <div className="relative group">
              <Database className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-green-900 group-focus-within:text-yellow-500 transition-colors" />
              <input 
                type="text" placeholder="FILTER BY VENDOR..."
                className="w-full bg-black border border-green-900 rounded-sm py-3 pl-10 pr-4 text-xs text-green-100 focus:border-yellow-500 outline-none transition-all placeholder:text-green-950 uppercase italic font-bold"
                value={filters.vendor}
                onChange={(e) => handleFilterChange('vendor', e.target.value)}
              />
            </div>
          </div>
        </div>
      </div>

      {isLoading ? (
        <div className="py-32 flex justify-center animate-pulse text-yellow-900 uppercase italic font-black tracking-widest text-xl">Indexing Inventory...</div>
      ) : isError ? (
        <div className="py-20 text-center text-red-500 border border-dashed border-red-900/30 font-black uppercase italic">Uplink Failure: Database Offline</div>
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
                    <th className="p-6">Asset Name</th>
                    <th className="p-6">Vendor</th>
                    <th className="p-6 text-right">Subscription</th>
                  </tr>
                </thead>
                <tbody className="text-sm">
                  {products.length === 0 ? (
                    <tr><td colSpan={3} className="py-24 text-center text-green-950 italic font-black uppercase tracking-widest">No Products Found</td></tr>
                  ) : products.map((p: any) => (
                    <tr key={p.id} className="border-b border-green-900/20 hover:bg-yellow-500/5 transition-all group">
                      <td className="p-6">
                        <div className="font-bold text-green-100 group-hover:text-yellow-500 uppercase tracking-tighter">{p.name}</div>
                        <div className="text-[9px] text-green-900 font-mono mt-1 uppercase tracking-widest">REF ID: {p.id.split('-')[0]}</div>
                      </td>
                      <td className="p-6 uppercase italic text-xs text-yellow-600 font-bold">{p.vendor_name}</td>
                      <td className="p-6 text-right">
                        {isAuthenticated ? (
                          <button 
                            onClick={() => p.subscription_id ? handleUnsubscribe(p.subscription_id, p.name) : handleSubscribe(p.id, p.name)}
                            disabled={actionLoading === p.id || actionLoading === p.subscription_id || isSubscribing || isUnsubscribing}
                            className={`px-4 py-2 text-[9px] font-black border transition-all disabled:opacity-50 disabled:cursor-not-allowed ${
                              p.subscription_id 
                                ? 'border-red-900 text-red-500 hover:bg-red-950/30' 
                                : 'border-green-500 text-green-500 hover:bg-green-500 hover:text-black'
                            }`}
                          >
                            {(actionLoading === p.id || actionLoading === p.subscription_id) ? (
                              <Loader2 className="w-3 h-3 animate-spin inline-block" />
                            ) : (
                              p.subscription_id ? 'UNSUBSCRIBE' : 'SUBSCRIBE'
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
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-4">
              {products.map((p: any) => (
                <div key={p.id} className="bg-black border border-green-900/50 p-5 hover:border-yellow-600 transition-all group relative overflow-hidden shadow-xl">
                  <div className="flex justify-between items-start mb-2">
                    <h3 className="font-bold text-green-100 group-hover:text-yellow-500 transition-colors uppercase truncate pr-4">{p.name || 'UNLABELED'}</h3>
                    <span className="text-[9px] text-green-900 font-mono italic">REF ID: {String(p?.id || '').split('-')[0] || 'N/A'}</span>
                  </div>
                  <div className="text-[10px] text-green-100 font-black mb-10 tracking-tighter uppercase italic opacity-80 font-sans font-bold">Origin: {p.vendor_name}</div>
                  {isAuthenticated ? (
                    <button 
                      onClick={() => p.subscription_id ? handleUnsubscribe(p.subscription_id, p.name) : handleSubscribe(p.id, p.name)}
                      disabled={actionLoading === p.id || actionLoading === p.subscription_id || isSubscribing || isUnsubscribing}
                      className={`w-full py-3 text-[10px] font-black border transition-all disabled:opacity-50 disabled:cursor-not-allowed ${
                        p.subscription_id 
                          ? 'border-red-900 text-red-500 hover:bg-red-950/30' 
                          : 'border-green-500 text-green-500 hover:bg-green-500 hover:text-black shadow-[0_0_10px_rgba(34,197,94,0.1)]'
                      }`}
                    >
                      {(actionLoading === p.id || actionLoading === p.subscription_id) ? (
                        <><Loader2 className="w-3 h-3 animate-spin inline-block mr-2" />PROCESSING...</>
                      ) : (
                        p.subscription_id ? 'UNSUBSCRIBE' : 'SUBSCRIBE PRODUCT'
                      )}
                    </button>
                  ) : (
                    <Link href="/login" className="block w-full py-3 text-center border border-dashed border-green-900/30 text-[9px] text-green-900 uppercase italic font-black hover:text-yellow-400 transition-colors">
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
