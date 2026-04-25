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

// src/theme/index.ts

import { colors } from './colors';
import { spacing } from './spacing';
import { typography } from './typography';

/**
 * theme is the unified design system object used by all dashboard components.
 */
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
