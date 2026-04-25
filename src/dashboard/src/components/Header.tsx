import React from 'react';
import { theme } from '@/theme';

const Header = () => {
  return (
    <header className="border-b sticky top-0 z-50 transition-all" style={{ 
      borderColor: theme.colors.border,
      backgroundColor: 'rgba(0,0,0,0.5)',
      backdropFilter: 'blur(12px)',
      height: theme.spacing.headerHeight 
    }}>
      <div className="mx-auto px-4 h-full flex items-center justify-between" style={{ maxWidth: theme.spacing.containerMax }}>
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: `linear-gradient(to bottom right, ${theme.colors.primary.DEFAULT}, #a855f7)` }}>
            <span className="text-white font-bold text-xl">G</span>
          </div>
          <h1 className="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-white to-white/60">
            GAIA Control Center
          </h1>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 px-3 py-1 rounded-full border" style={{ backgroundColor: theme.colors.success.subtle, borderColor: 'rgba(34, 197, 94, 0.2)' }}>
            <div className="w-2 h-2 rounded-full animate-pulse" style={{ backgroundColor: theme.colors.success.DEFAULT }} />
            <span className="text-xs font-medium tracking-wider uppercase" style={{ color: theme.colors.success.DEFAULT }}>Kernel Online</span>
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
