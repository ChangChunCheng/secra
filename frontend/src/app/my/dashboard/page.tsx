'use client';
import { useEffect, useState } from 'react';
import { useGetMeQuery, useUpdateMeMutation, useGetMyDashboardQuery, useUnsubscribeMutation, useUpdateThresholdMutation } from '@/lib/features/apiSlice';
import { Settings, Shield, X, Loader2, Mail, Globe, Bell, Save, CheckCircle2, Key, AlertCircle, ChevronLeft, ChevronRight } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useSelector } from 'react-redux';
import { RootState } from '@/lib/store';

export default function MyDashboard() {
  const [mounted, setMounted] = useState(false);
  const router = useRouter();
  
  const { data: user, isLoading: userLoading, isError: userError, error: userApiError } = useGetMeQuery();
  const { data: subs } = useGetMyDashboardQuery(undefined, { skip: !user });
  
  const [updateMe, { isLoading: updatingProfile }] = useUpdateMeMutation();
  const [unsubscribe] = useUnsubscribeMutation();
  const [updateThreshold] = useUpdateThresholdMutation();

  const [profileForm, setProfileForm] = useState({ email: '', timezone: '', notification_frequency: '', notification_time: '', password: '', confirmPassword: '' });
  const [showSuccess, setShowSuccess] = useState(false);
  const [formError, setFormError] = useState('');
  
  // Subscription list states
  const [activeTab, setActiveTab] = useState<'vendor' | 'product'>('vendor');
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  useEffect(() => { setMounted(true); }, []);

  useEffect(() => {
    if (mounted && !userLoading) {
      if (userError && (userApiError as any)?.status === 401) {
        router.push('/login');
      }
    }
  }, [mounted, userLoading, userError, userApiError, router]);

  useEffect(() => {
    if (user) {
      setProfileForm({
        email: user.email || '',
        timezone: user.timezone || 'UTC',
        notification_frequency: user.notification_frequency || 'daily',
        notification_time: user.notification_time || '09:00',
        password: '',
        confirmPassword: ''
      });
    }
  }, [user]);

  // Get current subscriptions based on active tab
  const currentSubs = activeTab === 'vendor' 
    ? (subs?.vendor_subs || []).map(sub => ({ ...sub, type: 'VENDOR' }))
    : (subs?.product_subs || []).map(sub => ({ ...sub, type: 'PRODUCT' }));
  
  // Pagination
  const totalPages = Math.ceil(currentSubs.length / itemsPerPage);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const paginatedSubs = currentSubs.slice(startIndex, endIndex);

  // Reset to page 1 when switching tabs
  useEffect(() => {
    setCurrentPage(1);
  }, [activeTab]);

  // Handle profile update
  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (profileForm.password && profileForm.password !== profileForm.confirmPassword) { setFormError('Passwords do not match'); return; }
    try {
      await updateMe({ 
        email: profileForm.email, 
        timezone: profileForm.timezone, 
        notification_frequency: profileForm.notification_frequency,
        notification_time: profileForm.notification_time,
        password: profileForm.password || undefined,
        confirm_password: profileForm.confirmPassword || undefined
      }).unwrap();
      setShowSuccess(true);
      setProfileForm({ ...profileForm, password: '', confirmPassword: '' });
      setTimeout(() => setShowSuccess(false), 3000);
    } catch (err: any) { setFormError(err.data?.error || 'Update failed'); }
  };

  if (!mounted || userLoading) return (
    <div className="min-h-screen bg-black flex items-center justify-center font-mono">
      <div className="flex flex-col items-center gap-4 animate-pulse">
        <Loader2 className="w-8 h-8 animate-spin text-green-500" />
        <span className="text-[10px] text-green-800 font-black uppercase tracking-widest">Verifying_Session...</span>
      </div>
    </div>
  );

  return (
    <div className="p-8 max-w-7xl mx-auto font-mono text-green-500">
      <div className="mb-16 border-b border-green-900 pb-10">
        <h1 className="text-4xl font-black text-green-400 tracking-tighter uppercase italic">{user?.username}&apos;s Interface</h1>
        <p className="text-[10px] text-green-800 tracking-[0.5em] mt-2 uppercase">Account Configuration</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-16">
        <div className="lg:col-span-1">
          <form onSubmit={handleUpdateProfile} className="bg-green-950/5 border border-green-900 p-6 rounded-sm space-y-6 shadow-2xl relative">
            {formError && <div className="p-3 bg-red-900/20 border border-red-900 text-red-500 text-[10px] font-black uppercase">{formError}</div>}
            
            <div>
              <label className="block text-[10px] text-green-800 uppercase font-black mb-2"><Mail className="inline w-3 h-3 mr-1" />Email_Address</label>
              <input type="email" className="w-full bg-black border border-green-900 rounded-sm px-4 py-2 text-sm text-green-100 outline-none focus:border-green-500" value={profileForm.email} onChange={(e) => setProfileForm({...profileForm, email: e.target.value})} required />
            </div>

            <div>
              <label className="block text-[10px] text-green-800 uppercase font-black mb-2"><Bell className="inline w-3 h-3 mr-1" />Notification_Frequency</label>
              <select className="w-full bg-black border border-green-900 rounded-sm px-4 py-2 text-sm text-green-100 outline-none focus:border-green-500 cursor-pointer" value={profileForm.notification_frequency} onChange={(e) => setProfileForm({...profileForm, notification_frequency: e.target.value})}>
                <option value="immediate">IMMEDIATE (On CVE Import)</option>
                <option value="daily">DAILY</option>
                <option value="weekly">WEEKLY</option>
              </select>
            </div>

            {(profileForm.notification_frequency === 'daily' || profileForm.notification_frequency === 'weekly') && (
              <div>
                <label className="block text-[10px] text-green-800 uppercase font-black mb-2">Notification_Time (HH:MM)</label>
                <input type="time" className="w-full bg-black border border-green-900 rounded-sm px-4 py-2 text-sm text-green-100 outline-none focus:border-green-500" value={profileForm.notification_time} onChange={(e) => setProfileForm({...profileForm, notification_time: e.target.value})} />
              </div>
            )}

            <div>
              <label className="block text-[10px] text-green-800 uppercase font-black mb-2"><Globe className="inline w-3 h-3 mr-1" />Timezone</label>
              <select className="w-full bg-black border border-green-900 rounded-sm px-4 py-2 text-sm text-green-100 outline-none focus:border-green-500 cursor-pointer" value={profileForm.timezone} onChange={(e) => setProfileForm({...profileForm, timezone: e.target.value})}>
                <option value="UTC">UTC</option>
                <option value="America/New_York">America/New_York (EST)</option>
                <option value="America/Los_Angeles">America/Los_Angeles (PST)</option>
                <option value="Europe/London">Europe/London (GMT)</option>
                <option value="Europe/Paris">Europe/Paris (CET)</option>
                <option value="Asia/Tokyo">Asia/Tokyo (JST)</option>
                <option value="Asia/Shanghai">Asia/Shanghai (CST)</option>
                <option value="Asia/Taipei">Asia/Taipei (TST)</option>
                <option value="Asia/Singapore">Asia/Singapore (SGT)</option>
                <option value="Australia/Sydney">Australia/Sydney (AEST)</option>
              </select>
            </div>

            <div className="border-t border-green-900/50 pt-6">
              <label className="block text-[10px] text-green-800 uppercase font-black mb-2"><Key className="inline w-3 h-3 mr-1" />New_Password (optional)</label>
              <input type="password" className="w-full bg-black border border-green-900 rounded-sm px-4 py-2 text-sm text-green-100 outline-none focus:border-green-500 mb-3" placeholder="Leave blank to keep current" value={profileForm.password} onChange={(e) => setProfileForm({...profileForm, password: e.target.value})} />
              
              <label className="block text-[10px] text-green-800 uppercase font-black mb-2">Confirm_Password</label>
              <input type="password" className="w-full bg-black border border-green-900 rounded-sm px-4 py-2 text-sm text-green-100 outline-none focus:border-green-500" placeholder="Leave blank to keep current" value={profileForm.confirmPassword} onChange={(e) => setProfileForm({...profileForm, confirmPassword: e.target.value})} />
            </div>

            <button type="submit" disabled={updatingProfile} className="w-full bg-green-500 hover:bg-green-400 text-black font-black py-3 rounded-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed">
              {updatingProfile ? 'UPDATING...' : 'SAVE_CHANGES'}
            </button>
            {showSuccess && <div className="absolute inset-0 bg-black/80 flex items-center justify-center font-black text-green-400 animate-in fade-in backdrop-blur-sm"><CheckCircle2 className="w-6 h-6 mr-2" />PROFILE_UPDATED</div>}
          </form>
        </div>

        <div className="lg:col-span-2 space-y-16">
          <section>
            <h2 className="text-sm font-black mb-6 flex items-center gap-3 text-green-400 uppercase"><Shield className="w-4 h-4" /> Active_Watches</h2>
            
            {/* Tab Navigation */}
            <div className="flex gap-2 mb-4">
              <button
                onClick={() => setActiveTab('vendor')}
                className={`px-6 py-2 text-[10px] font-black uppercase rounded-sm transition-all ${
                  activeTab === 'vendor'
                    ? 'bg-blue-900/30 text-blue-400 border-2 border-blue-900'
                    : 'bg-black text-green-800 border border-green-900 hover:text-green-500'
                }`}
              >
                Vendors ({subs?.vendor_subs?.length || 0})
              </button>
              <button
                onClick={() => setActiveTab('product')}
                className={`px-6 py-2 text-[10px] font-black uppercase rounded-sm transition-all ${
                  activeTab === 'product'
                    ? 'bg-purple-900/30 text-purple-400 border-2 border-purple-900'
                    : 'bg-black text-green-800 border border-green-900 hover:text-green-500'
                }`}
              >
                Products ({subs?.product_subs?.length || 0})
              </button>
            </div>
            
            <div className="border border-green-900 bg-green-950/5 rounded-sm overflow-hidden shadow-xl">
              <table className="w-full text-left">
                <thead className="text-[10px] text-green-800 uppercase border-b border-green-900 bg-green-900/10 font-black">
                  <tr><th className="p-5">Entity_Name</th><th className="p-5 w-32">Type</th><th className="p-5 w-64">Sensitivity</th><th className="p-5 w-20 text-center">Protocol</th></tr>
                </thead>
                <tbody className="text-sm">
                  {paginatedSubs.length === 0 ? (
                    <tr><td colSpan={4} className="p-12 text-center text-green-950 italic font-black uppercase">No active monitors detected</td></tr>
                  ) : paginatedSubs.map((sub: any) => (
                    <tr key={sub.id} className="border-b border-green-900/20 hover:bg-green-400/5 group">
                      <td className="p-5 font-bold text-green-100 group-hover:text-green-400 uppercase">{sub.target_name}</td>
                      <td className="p-5">
                        <span className={`inline-block px-3 py-1 text-[9px] font-black rounded-sm ${sub.type === 'VENDOR' ? 'bg-blue-900/30 text-blue-400 border border-blue-900' : 'bg-purple-900/30 text-purple-400 border border-purple-900'}`}>
                          {sub.type}
                        </span>
                      </td>
                      <td className="p-5">
                        <select value={sub.severity_threshold} onChange={(e) => updateThreshold({ id: sub.id, threshold: e.target.value })} className="bg-black border border-green-900 text-[10px] font-black text-green-400 p-2 rounded-sm w-full cursor-pointer outline-none">
                          {['INFO', 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL'].map(l => <option key={l} value={l}>{l}</option>)}
                        </select>
                      </td>
                      <td className="p-5 text-center">
                        <button onClick={() => unsubscribe(sub.id)} className="text-red-900 hover:text-red-500 font-black px-2 py-1 border border-red-900 transition-colors uppercase text-[10px]">Remove</button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              
              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-between p-4 border-t border-green-900 bg-green-900/5">
                  <div className="text-[10px] text-green-800 font-black uppercase">
                    Page {currentPage} of {totalPages} • {currentSubs.length} Total
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                      disabled={currentPage === 1}
                      className="px-3 py-1 bg-black border border-green-900 text-green-500 text-[10px] font-black uppercase rounded-sm hover:bg-green-900/20 disabled:opacity-30 disabled:cursor-not-allowed transition-all flex items-center gap-1"
                    >
                      <ChevronLeft className="w-3 h-3" /> Prev
                    </button>
                    <button
                      onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                      disabled={currentPage === totalPages}
                      className="px-3 py-1 bg-black border border-green-900 text-green-500 text-[10px] font-black uppercase rounded-sm hover:bg-green-900/20 disabled:opacity-30 disabled:cursor-not-allowed transition-all flex items-center gap-1"
                    >
                      Next <ChevronRight className="w-3 h-3" />
                    </button>
                  </div>
                </div>
              )}
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
