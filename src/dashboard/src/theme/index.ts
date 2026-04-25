// src/theme/index.ts

import { colors } from './colors';
import { spacing } from './spacing';
import { typography } from './typography';

export const theme = {
  colors,
  spacing,
  typography,
  radius: {
    xs: '0.25rem',
    sm: '0.5rem',
    md: '0.75rem',
    lg: '1rem',
    xl: '1.25rem',
    '2xl': '1.5rem',
    full: '9999px',
  },
  shadows: {
    glow: `0 0 20px -5px ${colors.primary.glow}`,
    success: `0 0 15px -3px ${colors.success.glow}`,
  }
};

export * from './colors';
export * from './spacing';
export * from './typography';
