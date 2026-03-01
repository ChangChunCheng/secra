'use client';
import { Inter } from 'next/font/google';
import './globals.css';
import StoreProvider from './StoreProvider';
import Navbar from '@/components/Navbar';
import AuthInit from '@/components/AuthInit';
import { useState, useEffect } from 'react';

const inter = Inter({ subsets: ['latin'] });

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const [mounted, setMounted] = useState(false);
  const [error, setError] = useState<any>(null);

  useEffect(() => {
    setMounted(true);
    const handleError = (e: ErrorEvent) => {
      console.error('Captured Runtime Error:', e.error);
      setError(e.error);
    };
    window.addEventListener('error', handleError);
    return () => window.removeEventListener('error', handleError);
  }, []);

  return (
    <html lang="en">
      <body className={`${inter.className} bg-black min-h-screen text-green-500 custom-scrollbar`}>
        <StoreProvider>
          <AuthInit>
            {!mounted ? (
              <div className="fixed inset-0 bg-black flex items-center justify-center font-mono">
                <span className="text-green-900 animate-pulse uppercase tracking-[0.5em]">Initializing_Grid...</span>
              </div>
            ) : error ? (
              <div className="min-h-screen flex items-center justify-center p-8 bg-black font-mono">
                <div className="max-w-2xl border border-red-500 p-10 shadow-[0_0_50px_rgba(239,68,68,0.1)]">
                  <h1 className="text-red-500 text-3xl font-black mb-4 uppercase italic">Fatal_Client_Exception</h1>
                  <p className="text-xs text-red-900 mb-8 leading-relaxed uppercase">
                    A critical synchronization error has occurred in the reactive layer. 
                    Please reset the interface or clear local cache.
                  </p>
                  <pre className="bg-red-900/10 p-4 border border-red-900/30 text-[10px] text-red-400 overflow-auto mb-10 max-h-40">
                    {error.stack || error.message || 'Stack trace encrypted'}
                  </pre>
                  <button 
                    onClick={() => window.location.reload()}
                    className="w-full py-3 bg-red-500 text-black font-black uppercase tracking-widest hover:bg-red-400 transition-all"
                  >
                    Reset_Grid_Interface
                  </button>
                </div>
              </div>
            ) : (
              <>
                <Navbar />
                <main className="transition-opacity duration-500 ease-in-out opacity-100">
                  {children}
                </main>
              </>
            )}
          </AuthInit>
        </StoreProvider>
      </body>
    </html>
  );
}
