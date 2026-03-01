'use client';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useSelector, useDispatch } from 'react-redux';
import { RootState } from '@/lib/store';
import { logout as clearLocalAuth } from '@/lib/features/authSlice';
import { useLogoutApiMutation } from '@/lib/features/apiSlice';
import { Shield, LayoutDashboard, Database, Package, User, LogOut, ShieldCheck } from 'lucide-react';

export default function Navbar() {
  const pathname = usePathname();
  const router = useRouter();
  const dispatch = useDispatch();
  const { isAuthenticated, user } = useSelector((state: RootState) => state.auth);
  const [logoutApi] = useLogoutApiMutation();

  const navItems = [
    { name: 'Dashboard', path: '/', icon: LayoutDashboard },
    { name: 'CVEs', path: '/cves', icon: Shield },
    { name: 'Vendors', path: '/vendors', icon: Database },
    { name: 'Products', path: '/products', icon: Package },
  ];

  const handleLogout = async () => {
    try {
      await logoutApi().unwrap(); 
    } catch (e) {
      console.error('Logout API failed', e);
    } finally {
      dispatch(clearLocalAuth()); 
      if (pathname.startsWith('/my') || pathname.startsWith('/admin')) {
        router.push('/login');
      }
    }
  };

  return (
    <nav className="bg-black border-b border-green-900 sticky top-0 z-50 shadow-[0_4px_20px_rgba(0,0,0,0.8)]">
      <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between font-mono">
        <div className="flex items-center gap-10">
          <Link href="/" className="flex items-center gap-3 group">
            <Shield className="w-7 h-7 text-green-500 group-hover:rotate-12 transition-transform" />
            <span className="text-2xl font-black italic tracking-tighter text-green-400 group-hover:text-green-300 uppercase">SECRA</span>
          </Link>

          <div className="hidden md:flex items-center gap-6">
            {navItems.map((item) => {
              const Icon = item.icon;
              const isActive = pathname === item.path;
              return (
                <Link 
                  key={item.path} 
                  href={item.path}
                  className={`flex items-center gap-2 text-[11px] font-black uppercase tracking-widest transition-all ${
                    isActive ? 'text-green-400' : 'text-green-900 hover:text-green-500'
                  }`}
                >
                  <Icon className="w-3.5 h-3.5" /> {item.name}
                </Link>
              );
            })}
          </div>
        </div>

        <div className="flex items-center gap-6">
          {isAuthenticated ? (
            <>
              {user?.role === 'admin' && (
                <Link href="/admin/users" className="text-yellow-600 hover:text-yellow-400 flex items-center gap-2 text-[10px] font-black uppercase tracking-tighter">
                  <ShieldCheck className="w-4 h-4" /> Admin
                </Link>
              )}
              <Link href="/my/dashboard" className="text-green-100 hover:text-green-400 flex items-center gap-2 text-[10px] font-black uppercase group">
                <User className="w-4 h-4 text-green-500" /> 
                <span className="border-b border-dashed border-green-900 group-hover:border-green-400 transition-colors">
                  {user?.username || 'Profile'}
                </span>
              </Link>
              <button 
                onClick={handleLogout}
                className="flex items-center gap-2 text-red-900 hover:text-red-500 text-[10px] font-black uppercase"
              >
                <LogOut className="w-4 h-4" /> Sign Out
              </button>
            </>
          ) : (
            <div className="flex gap-4">
              <Link href="/login" className="text-green-900 hover:text-green-400 text-[10px] font-black uppercase tracking-widest border border-green-900 px-4 py-1.5 rounded-sm">Sign In</Link>
              <Link href="/register" className="bg-green-500 hover:bg-green-400 text-black text-[10px] font-black uppercase tracking-widest px-4 py-1.5 rounded-sm shadow-lg shadow-green-900/20">Register</Link>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
}
