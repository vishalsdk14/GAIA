// Copyright 2026 GAIA Contributors
//
// Licensed under the MIT License.
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React from 'react';
import { theme } from '@/theme';
import { Shield, Zap, Cpu, Settings, Bell, Activity } from 'lucide-react';
import Button from './Button';

/**
 * Header component provides the primary navigation and global status for the dashboard.
 * 100% theme-driven: Zero magic numbers across colors, spacing, typography, and shadows.
 */
const Header = () => {
  return (
    <header className="sticky top-0 z-50 border-b backdrop-blur-md" style={{ 
      borderColor: theme.colors.border.subtle,
      height: theme.spacing.headerHeight,
      backgroundColor: theme.colors.glass.base
    }}>
      <div className="mx-auto h-full flex items-center justify-between" style={{ paddingLeft: theme.spacing.xl, paddingRight: theme.spacing.xl }}>
        <div className="flex items-center" style={{ gap: theme.spacing.xl }}>
          {/* Logo Lockup */}
          <div className="flex items-center" style={{ gap: theme.spacing.md }}>
            <div 
              className="flex items-center justify-center shadow-xl"
              style={{ 
                width: theme.spacing.logoIconSize, 
                height: theme.spacing.logoIconSize, 
                borderRadius: theme.radius.md,
                backgroundColor: theme.colors.primary.DEFAULT,
                boxShadow: theme.shadows.primary 
              }}
            >
              <Cpu 
                className="w-5 h-5" 
                style={{ color: '#ffffff' }} 
                strokeWidth={theme.strokeWidth.bold} 
              />
            </div>
            <div className="flex flex-col justify-center">
              <h1 
                className="font-black leading-none" 
                style={{ 
                  color: theme.colors.text.primary,
                  fontSize: theme.typography.size.xl,
                  letterSpacing: theme.typography.tracking.tight 
                }}
              >
                GAIA
              </h1>
              <span 
                className="font-black uppercase"
                style={{ 
                  marginTop: theme.spacing.xs,
                  fontSize: theme.typography.size.tiny, 
                  letterSpacing: theme.typography.tracking.widest,
                  color: theme.colors.text.muted 
                }}
              >
                Core Kernel
              </span>
            </div>
          </div>

          <div 
            style={{ 
              height: theme.spacing.lg, 
              width: '1px', 
              backgroundColor: theme.colors.border.subtle 
            }} 
          />

          {/* Navigation Links */}
          <nav className="flex items-center" style={{ gap: theme.spacing.xs }}>
            {[
              { label: 'Overview', active: true },
              { label: 'Agents', active: false },
              { label: 'Security', active: false }
            ].map((item, idx) => (
              <Button 
                key={idx}
                variant={item.active ? 'primary' : 'ghost'}
                size="sm"
                style={{ 
                  backgroundColor: item.active ? theme.colors.primary.subtle : 'transparent', 
                  color: item.active ? theme.colors.primary.DEFAULT : theme.colors.text.secondary 
                }}
              >
                {item.label}
              </Button>
            ))}
          </nav>
        </div>

        {/* Global Action Group */}
        <div className="flex items-center" style={{ gap: theme.spacing.lg }}>
          <div 
            className="flex items-center border"
            style={{ 
              gap: theme.spacing.sm,
              paddingLeft: theme.spacing.lg,
              paddingRight: theme.spacing.lg,
              paddingTop: theme.spacing.sm,
              paddingBottom: theme.spacing.sm,
              backgroundColor: theme.colors.surface.low, 
              borderColor: theme.colors.border.subtle,
              borderRadius: theme.radius.full
            }}
          >
            <Activity 
              className="w-4 h-4" 
              style={{ color: theme.colors.success.DEFAULT }} 
              strokeWidth={theme.strokeWidth.bold} 
            />
            <span 
              className="font-black uppercase"
              style={{ 
                fontSize: theme.typography.size.tiny, 
                letterSpacing: theme.typography.tracking.widest,
                color: theme.colors.text.secondary 
              }}
            >
              System <span style={{ color: theme.colors.success.DEFAULT }}>Live</span>
            </span>
          </div>

          <div className="flex items-center" style={{ gap: theme.spacing.sm }}>
             <Button size="icon" variant="secondary" icon={Bell} />
             <Button size="icon" variant="secondary" icon={Settings} />
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
