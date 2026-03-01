'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import axios from 'axios';
import { UserPlus, AlertCircle, Loader2 } from 'lucide-react';

export default function RegisterPage() {
  const [mounted, setMounted] = useState(false);
  const [form, setForm] = useState({ username: '', email: '', password: '' });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  useEffect(() => { setMounted(true); }, []);

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await axios.post('/api/v1/auth/register', form);
      router.push('/login');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  if (!mounted) return <div className="bg-black min-h-screen" />;

  return (
    <div className="min-h-[80vh] flex items-center justify-center p-6 font-mono text-green-500">
      <div className="w-full max-w-md border border-green-900 bg-green-950/5 p-8 rounded-sm">
        <div className="mb-10 text-center">
          <UserPlus className="w-12 h-12 text-green-500 mx-auto mb-4" />
          <h2 className="text-2xl font-black italic tracking-tighter uppercase">Protocol_Enrollment</h2>
          <p className="text-[10px] text-green-800 uppercase mt-1 tracking-widest">Create New Identity Node</p>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-900/20 border border-red-900 text-red-400 text-xs flex items-center gap-3 italic">
            <AlertCircle className="w-4 h-4" /> ERROR: {error}
          </div>
        )}

        <form onSubmit={handleRegister} className="space-y-6">
          <div>
            <label className="block text-[10px] text-green-800 uppercase font-black mb-2 tracking-widest">Username</label>
            <input 
              type="text" required
              className="w-full bg-black border border-green-900 rounded-sm px-4 py-3 text-sm focus:border-green-400 outline-none transition-all"
              value={form.username}
              onChange={(e) => setForm({...form, username: e.target.value})}
            />
          </div>
          <div>
            <label className="block text-[10px] text-green-800 uppercase font-black mb-2 tracking-widest">Email_Address</label>
            <input 
              type="email" required
              className="w-full bg-black border border-green-900 rounded-sm px-4 py-3 text-sm focus:border-green-400 outline-none transition-all"
              value={form.email}
              onChange={(e) => setForm({...form, email: e.target.value})}
            />
          </div>
          <div>
            <label className="block text-[10px] text-green-800 uppercase font-black mb-2 tracking-widest">Secure_Password</label>
            <input 
              type="password" required
              className="w-full bg-black border border-green-900 rounded-sm px-4 py-3 text-sm focus:border-green-400 outline-none transition-all"
              value={form.password}
              onChange={(e) => setForm({...form, password: e.target.value})}
            />
          </div>
          <button 
            type="submit" disabled={loading}
            className="w-full bg-green-500 hover:bg-green-400 text-black font-black py-3 rounded-sm transition-all shadow-xl active:scale-[0.98] flex justify-center"
          >
            {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : 'EXECUTE_ENROLLMENT'}
          </button>
        </form>
      </div>
    </div>
  );
}
