'use client';
import { useState } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useSelector, useDispatch } from 'react-redux';
import { RootState } from '@/lib/store';
import { logout as clearLocalAuth } from '@/lib/features/authSlice';
import { useLogoutApiMutation } from '@/lib/features/apiSlice';
import { Shield, LayoutDashboard, Database, Package, User, LogOut, ShieldCheck, Menu, X } from 'lucide-react';

export default function Navbar() {
  const pathname = usePathname();
  const router = useRouter();
  const dispatch = useDispatch();
  const { isAuthenticated, user } = useSelector((state: RootState) => state.auth);
  const [logoutApi] = useLogoutApiMutation();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

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
      setMobileMenuOpen(false); // Close mobile menu on logout
      if (pathname.startsWith('/my') || pathname.startsWith('/admin')) {
        router.push('/login');
      }
    }
  };

  const closeMobileMenu = () => setMobileMenuOpen(false);

  return (
    <>
      <nav className="bg-black border-b border-green-900 sticky top-0 z-[100] shadow-[0_4px_20px_rgba(0,0,0,0.8)]">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between font-mono">
          <div className="flex items-center gap-10">
            <Link href="/" className="flex items-center gap-3 group">
              <Shield className="w-7 h-7 text-green-500 group-hover:rotate-12 transition-transform" />
              <span className="text-2xl font-black italic tracking-tighter text-green-400 group-hover:text-green-300 uppercase">SECRA</span>
            </Link>

            {/* Desktop Navigation */}
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

          {/* Desktop Auth Buttons */}
          <div className="hidden md:flex items-center gap-6">
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

          {/* Mobile Menu Button */}
          <button 
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="md:hidden text-green-500 hover:text-green-400 transition-colors"
            aria-label="Toggle menu"
          >
            {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>
      </nav>

      {/* Mobile Menu Overlay */}
      {mobileMenuOpen && (
        <div className="fixed inset-0 z-[90] md:hidden">
          {/* Backdrop */}
          <div 
            className="absolute inset-0 bg-black/95 backdrop-blur-sm"
            onClick={closeMobileMenu}
          />
          
          {/* Menu Panel */}
          <div className="relative bg-black border-l border-green-900 ml-auto w-4/5 h-full shadow-2xl font-mono">
            <div className="flex flex-col h-full">
              {/* Mobile Navigation Links */}
              <div className="flex-1 overflow-y-auto p-6 space-y-2">
                <div className="text-[10px] text-green-800 font-black uppercase mb-6 tracking-widest">Navigation</div>
                {navItems.map((item) => {
                  const Icon = item.icon;
                  const isActive = pathname === item.path;
                  return (
                    <Link
                      key={item.path}
                      href={item.path}
                      onClick={closeMobileMenu}
                      className={`flex items-center gap-3 px-4 py-3 rounded-sm text-sm font-black uppercase tracking-wider transition-all ${
                        isActive 
                          ? 'bg-green-500/10 text-green-400 border border-green-500/50' 
                          : 'text-green-100 hover:bg-green-900/20 hover:text-green-400'
                      }`}
                    >
                      <Icon className="w-5 h-5" /> {item.name}
                    </Link>
                  );
                })}
              </div>

              {/* Mobile Auth Section */}
              <div className="border-t border-green-900 p-6 bg-green-950/20">
                {isAuthenticated ? (
                  <div className="space-y-3">
                    <div className="text-[10px] text-green-800 font-black uppercase mb-3 tracking-widest">Account</div>
                    
                    {user?.role === 'admin' && (
                      <Link 
                        href="/admin/users"
                        onClick={closeMobileMenu}
                        className="flex items-center gap-3 px-4 py-3 rounded-sm text-sm font-black uppercase bg-yellow-900/20 text-yellow-500 border border-yellow-900"
                      >
                        <ShieldCheck className="w-5 h-5" /> Admin Panel
                      </Link>
                    )}
                    
                    <Link 
                      href="/my/dashboard"
                      onClick={closeMobileMenu}
                      className="flex items-center gap-3 px-4 py-3 rounded-sm text-sm font-black uppercase text-green-100 hover:bg-green-900/20 hover:text-green-400 transition-all"
                    >
                      <User className="w-5 h-5" /> {user?.username || 'My Profile'}
                    </Link>
                    
                    <button
                      onClick={handleLogout}
                      className="w-full flex items-center gap-3 px-4 py-3 rounded-sm text-sm font-black uppercase bg-red-900/20 text-red-500 border border-red-900 hover:bg-red-900/30 transition-all"
                    >
                      <LogOut className="w-5 h-5" /> Sign Out
                    </button>
                  </div>
                ) : (
                  <div className="space-y-3">
                    <Link 
                      href="/login"
                      onClick={closeMobileMenu}
                      className="block text-center px-4 py-3 rounded-sm text-sm font-black uppercase border border-green-900 text-green-400 hover:bg-green-900/20 transition-all"
                    >
                      Sign In
                    </Link>
                    <Link 
                      href="/register"
                      onClick={closeMobileMenu}
                      className="block text-center px-4 py-3 rounded-sm text-sm font-black uppercase bg-green-500 text-black hover:bg-green-400 transition-all"
                    >
                      Register
                    </Link>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
