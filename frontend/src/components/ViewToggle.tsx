'use client';
import React from 'react';
import { LayoutList, LayoutGrid } from 'lucide-react';

export type ViewMode = 'list' | 'card';

interface ViewToggleProps {
  mode: ViewMode;
  onModeChange: (mode: ViewMode) => void;
}

export default function ViewToggle({ mode, onModeChange }: ViewToggleProps) {
  return (
    <div className="flex border border-green-900 rounded-sm bg-black overflow-hidden shadow-lg">
      <button 
        onClick={() => onModeChange('list')}
        className={`flex items-center gap-2 px-4 py-2 text-[10px] font-black transition-all ${
          mode === 'list' ? 'bg-green-500 text-black' : 'text-green-900 hover:text-green-400'
        }`}
      >
        <LayoutList className="w-3.5 h-3.5" /> LIST
      </button>
      <button 
        onClick={() => onModeChange('card')}
        className={`flex items-center gap-2 px-4 py-2 text-[10px] font-black transition-all ${
          mode === 'card' ? 'bg-green-500 text-black' : 'text-green-900 hover:text-green-400'
        }`}
      >
        <LayoutGrid className="w-3.5 h-3.5" /> CARDS
      </button>
    </div>
  );
}
