'use client';
import { useState } from 'react';
import { useDispatch } from 'react-redux';
import { useRouter } from 'next/navigation';
import { setCredentials } from '@/lib/features/authSlice';
import axios from 'axios';
import { LogIn, AlertCircle, Loader2 } from 'lucide-react';

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const router = useRouter();
  const dispatch = useDispatch();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      // 1. Initial Login to get session cookie
      await axios.post('/api/v1/auth/login', { username, password });
      
      // 2. Fetch full user profile (with role)
      const userRes = await axios.get('/api/v1/me');
      
      // 3. Dispatch full user data to Redux
      dispatch(setCredentials(userRes.data));
      
      router.push('/');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Authentication failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[80vh] flex items-center justify-center p-6 font-mono">
      <div className="w-full max-w-md bg-green-950/5 border border-green-900 p-8 rounded-sm shadow-2xl">
        <div className="mb-10 text-center">
          <LogIn className="w-12 h-12 text-green-500 mx-auto mb-4" />
          <h2 className="text-2xl font-black text-green-400 italic tracking-tighter uppercase">Protocol_Login</h2>
          <p className="text-[10px] text-green-800 uppercase mt-1 tracking-widest font-bold">Input Credentials to Initialize Session</p>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-900/20 border border-red-900 text-red-400 text-xs flex items-center gap-3 italic">
            <AlertCircle className="w-4 h-4" /> ERROR: {error}
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-6">
          <div>
            <label className="block text-[10px] text-green-800 uppercase font-black mb-2 tracking-widest">Username</label>
            <input 
              type="text" 
              required
              className="w-full bg-black border border-green-900 rounded-sm px-4 py-3 text-sm text-green-100 focus:border-green-400 outline-none transition-all"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-[10px] text-green-800 uppercase font-black mb-2 tracking-widest">Password</label>
            <input 
              type="password" 
              required
              className="w-full bg-black border border-green-900 rounded-sm px-4 py-3 text-sm text-green-100 focus:border-green-400 outline-none transition-all"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          <button 
            type="submit" 
            disabled={loading}
            className="w-full bg-green-500 hover:bg-green-400 disabled:opacity-30 text-black font-black py-3 rounded-sm transition-all shadow-lg active:scale-95 flex justify-center"
          >
            {loading ? <Loader2 className="w-5 h-5 animate-spin" /> : 'INITIALIZE_LOGIN'}
          </button>
        </form>
      </div>
    </div>
  );
}
