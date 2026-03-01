'use client';
import React, { useState, useEffect } from 'react';
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';

interface PaginationProps {
  currentPage: number;
  totalPages: number;
  totalItems: number;
  onPageChange: (page: number) => void;
}

export default function Pagination({ currentPage, totalPages, totalItems, onPageChange }: PaginationProps) {
  const [jumpPage, setJumpPage] = useState(currentPage.toString());

  useEffect(() => {
    setJumpPage(currentPage.toString());
  }, [currentPage]);

  const handleJump = (e: React.FormEvent) => {
    e.preventDefault();
    const p = parseInt(jumpPage);
    if (!isNaN(p) && p >= 1 && p <= totalPages) {
      onPageChange(p);
    } else {
      setJumpPage(currentPage.toString());
    }
  };

  if (totalPages <= 1) return (
    <div className="p-4 bg-green-950/5 border-t border-green-900/30 text-[10px] text-green-800 uppercase font-black">
      Total Records: {totalItems.toLocaleString()}
    </div>
  );

  return (
    <div className="p-4 bg-green-950/10 border-t border-green-900 flex flex-col md:flex-row justify-between items-center gap-4">
      <div className="text-[10px] text-green-800 uppercase font-black">
        Total Found: <span className="text-green-500">{totalItems.toLocaleString()}</span>
      </div>

      <div className="flex items-center gap-2">
        <div className="flex gap-1 mr-4">
          <button 
            disabled={currentPage === 1}
            onClick={() => onPageChange(1)}
            className="p-2 border border-green-900 text-green-900 hover:border-green-400 disabled:opacity-10 transition-all"
          >
            <ChevronsLeft className="w-4 h-4" />
          </button>
          <button 
            disabled={currentPage === 1}
            onClick={() => onPageChange(currentPage - 1)}
            className="p-2 border border-green-900 text-green-900 hover:border-green-400 disabled:opacity-10 transition-all"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
        </div>

        <span className="text-[10px] font-black tracking-widest text-green-400 uppercase">
          Page {currentPage} of {totalPages}
        </span>

        <div className="flex gap-1 ml-4">
          <button 
            disabled={currentPage === totalPages}
            onClick={() => onPageChange(currentPage + 1)}
            className="p-2 border border-green-900 text-green-900 hover:border-green-400 disabled:opacity-10 transition-all"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
          <button 
            disabled={currentPage === totalPages}
            onClick={() => onPageChange(totalPages)}
            className="p-2 border border-green-900 text-green-900 hover:border-green-400 disabled:opacity-10 transition-all"
          >
            <ChevronsRight className="w-4 h-4" />
          </button>
        </div>
      </div>

      <form onSubmit={handleJump} className="flex items-center gap-2">
        <span className="text-[9px] text-green-900 uppercase font-bold tracking-tighter">Go to page:</span>
        <input 
          type="text"
          className="w-12 bg-black border border-green-900 rounded-sm py-1 px-2 text-[10px] text-green-400 focus:border-green-400 outline-none text-center"
          value={jumpPage}
          onChange={(e) => setJumpPage(e.target.value)}
        />
      </form>
    </div>
  );
}
