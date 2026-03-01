'use client';
import { useEffect, useState } from 'react';
import { useGetAdminUsersQuery, useUpdateUserRoleMutation } from '@/lib/features/apiSlice';
import { ShieldCheck, User, Loader2, AlertCircle, CheckCircle2 } from 'lucide-react';
import { useSelector } from 'react-redux';
import { RootState } from '@/lib/store';
import { useRouter } from 'next/navigation';

export default function AdminUsersPage() {
  const [mounted, setMounted] = useState(false);
  const { user: currentUser, isAuthenticated } = useSelector((state: RootState) => state.auth);
  const router = useRouter();
  
  const { data: users, isLoading, isError, refetch } = useGetAdminUsersQuery();
  const [updateUserRole, { isLoading: updating }] = useUpdateUserRoleMutation();
  const [successID, setSuccessID] = useState<string | null>(null);

  useEffect(() => {
    setMounted(true);
    // CRITICAL: Immediate redirect if not logged in OR not admin
    if (mounted && !isAuthenticated) {
      router.push('/login');
    } else if (mounted && currentUser && currentUser.role !== 'admin') {
      router.push('/');
    }
  }, [currentUser, isAuthenticated, mounted, router]);

  const handleRoleChange = async (userID: string, newRole: string) => {
    try {
      await updateUserRole({ user_id: userID, role: newRole }).unwrap();
      setSuccessID(userID);
      setTimeout(() => setSuccessID(null), 2000);
      refetch();
    } catch (err) {
      console.error('Failed to update role', err);
    }
  };

  if (!mounted || !isAuthenticated || currentUser?.role !== 'admin') return <div className="bg-black min-h-screen" />;

  if (isError) return (
    <div className="p-20 text-center text-red-500 border border-dashed border-red-900 mx-auto max-w-7xl mt-10 font-mono shadow-2xl">
      <AlertCircle className="mx-auto mb-4 w-10 h-10" />
      <h2 className="text-2xl font-black mb-2 uppercase italic">Clearance Required</h2>
      <p className="text-xs uppercase opacity-70">Administrator access is required for this area.</p>
    </div>
  );

  return (
    <div className="p-8 max-w-7xl mx-auto font-mono text-green-500">
      <div className="mb-12 border-l-4 border-yellow-500 pl-6 py-2 bg-yellow-500/5 flex justify-between items-end">
        <div>
          <h1 className="text-3xl font-black text-yellow-500 uppercase italic tracking-tighter flex items-center gap-3">
            <ShieldCheck className="w-8 h-8" /> Admin Panel
          </h1>
          <p className="text-[10px] text-yellow-800 uppercase tracking-widest mt-1">User Management & Privilege Control</p>
        </div>
        <div className="text-right">
          <div className="text-[9px] text-yellow-900 font-black mb-1 uppercase tracking-widest">Registered Users</div>
          <div className="text-2xl font-bold text-yellow-500 tabular-nums">{users?.length || 0}</div>
        </div>
      </div>

      <div className="bg-black border border-green-900 rounded-sm overflow-hidden shadow-2xl relative">
        {updating && (
          <div className="absolute inset-0 bg-black/40 backdrop-blur-[1px] z-10 flex items-center justify-center">
            <Loader2 className="w-8 h-8 animate-spin text-yellow-500" />
          </div>
        )}
        
        <table className="w-full text-left">
          <thead className="bg-green-900/10 text-green-800 text-[10px] uppercase font-black border-b border-green-900">
            <tr>
              <th className="p-5">Username</th>
              <th className="p-5">Email Address</th>
              <th className="p-5">Role</th>
              <th className="p-5 text-right">Join Date</th>
            </tr>
          </thead>
          <tbody className="text-sm">
            {isLoading ? (
              <tr><td colSpan={4} className="py-20 text-center animate-pulse text-green-900 uppercase italic tracking-widest">Accessing Directory...</td></tr>
            ) : users?.map((u: any) => (
              <tr key={u.id} className="border-b border-green-900/20 hover:bg-yellow-500/5 transition-colors group">
                <td className="p-5 flex items-center gap-4">
                  <div className={`w-8 h-8 rounded-sm flex items-center justify-center transition-all ${
                    u.role === 'admin' 
                      ? 'bg-yellow-500 text-black shadow-[0_0_15px_rgba(234,179,8,0.4)]' 
                      : 'bg-green-900/20 text-green-500 border border-green-900'
                  }`}>
                    {successID === u.id ? <CheckCircle2 className="w-4 h-4 animate-in zoom-in duration-300" /> : <User className="w-4 h-4" />}
                  </div>
                  <span className="font-bold text-green-100 uppercase tracking-tighter">{u.username}</span>
                </td>
                <td className="p-5 text-green-800 font-sans text-xs tracking-tight">{u.email}</td>
                <td className="p-5">
                  <select 
                    value={u.role}
                    disabled={u.id === currentUser?.id}
                    onChange={(e) => handleRoleChange(u.id, e.target.value)}
                    className={`bg-black border rounded-sm px-3 py-1.5 text-[10px] font-black focus:outline-none transition-all cursor-pointer uppercase disabled:opacity-30 disabled:cursor-not-allowed ${
                      u.role === 'admin' ? 'border-yellow-600 text-yellow-500' : 'border-green-900 text-green-500'
                    }`}
                  >
                    <option value="user">USER</option>
                    <option value="admin">ADMIN</option>
                  </select>
                </td>
                <td className="p-5 text-right text-[10px] text-green-900 font-mono italic">
                  {new Date(u.created_at).toISOString().split('T')[0]}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
