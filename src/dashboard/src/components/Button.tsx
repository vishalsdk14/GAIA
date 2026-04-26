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

'use client';

import React from 'react';
import { theme } from '@/theme';
import { LucideIcon } from 'lucide-react';

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger' | 'success';
type ButtonSize = 'sm' | 'md' | 'lg' | 'icon' | 'icon-sm' | 'icon-lg';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  icon?: LucideIcon;
  isLoading?: boolean;
}

/**
 * Button is a modular UI component that enforces consistent styling across variants.
 * It uses theme tokens for all categories: colors, spacing, typography, and shadows.
 */
const Button = ({ 
  children, 
  variant = 'primary', 
  size = 'md', 
  icon: Icon, 
  isLoading, 
  className = '', 
  style, 
  ...props 
}: ButtonProps) => {
  
  const getVariantStyles = () => {
    switch (variant) {
      case 'primary':
        return {
          backgroundColor: theme.colors.primary.DEFAULT,
          color: '#ffffff',
          boxShadow: theme.shadows.primary,
          border: 'none',
        };
      case 'secondary':
        return {
          backgroundColor: '#ffffff',
          color: theme.colors.text.secondary,
          borderColor: theme.colors.border.subtle,
          border: `1px solid ${theme.colors.border.subtle}`,
        };
      case 'ghost':
        return {
          backgroundColor: 'transparent',
          color: theme.colors.text.muted,
          border: 'none',
        };
      case 'danger':
        return {
          backgroundColor: theme.colors.error.subtle,
          color: theme.colors.error.DEFAULT,
          border: 'none',
        };
      case 'success':
        return {
          backgroundColor: theme.colors.success.subtle,
          color: theme.colors.success.DEFAULT,
          border: 'none',
        };
      default:
        return {};
    }
  };

  const getSizeStyles = () => {
    switch (size) {
      case 'sm':
        return { 
          padding: `${theme.spacing.xs} ${theme.spacing.md}`, 
          fontSize: theme.typography.size.xs 
        };
      case 'md':
        return { 
          padding: `${theme.spacing.sm} ${theme.spacing.lg}`, 
          fontSize: theme.typography.size.sm 
        };
      case 'lg':
        return { 
          padding: `${theme.spacing.md} ${theme.spacing.xl}`, 
          fontSize: theme.typography.size.base 
        };
      case 'icon-sm':
        return { width: theme.spacing.iconButtonSm, height: theme.spacing.iconButtonSm, padding: '0' };
      case 'icon':
        return { width: theme.spacing.iconButtonMd, height: theme.spacing.iconButtonMd, padding: '0' };
      case 'icon-lg':
        return { width: theme.spacing.iconButtonLg, height: theme.spacing.iconButtonLg, padding: '0' };
      default:
        return {};
    }
  };

  const baseStyles: React.CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: theme.spacing.sm,
    fontWeight: theme.typography.weight.black,
    textTransform: 'uppercase',
    letterSpacing: theme.typography.tracking.wider,
    transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)',
    borderRadius: theme.radius.md,
    cursor: props.disabled || isLoading ? 'not-allowed' : 'pointer',
    opacity: props.disabled || isLoading ? 0.6 : 1,
    ...getVariantStyles(),
    ...getSizeStyles(),
    ...style,
  };

  return (
    <button 
      className={`group active:scale-95 hover:brightness-110 ${className}`}
      style={baseStyles}
      {...props}
    >
      {isLoading ? (
        <div 
          className="border-2 border-current border-t-transparent rounded-full animate-spin" 
          style={{ width: theme.spacing.md, height: theme.spacing.md }}
        />
      ) : (
        <>
          {Icon && (
            <Icon 
              className={`${size === 'sm' || size === 'icon-sm' ? 'w-3.5 h-3.5' : 'w-4 h-4'}`} 
              strokeWidth={2.5}
            />
          )}
          {children}
        </>
      )}
    </button>
  );
};

export default Button;
