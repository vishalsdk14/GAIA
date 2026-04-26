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

import { colors } from './colors';

/**
 * shadows defines the elevation and glow system for the dashboard.
 */
export const shadows = {
  sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
  md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
  lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
  xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
  
  // Brand Glows
  primary: `0 8px 16px -4px ${colors.primary.glow}`,
  primaryLarge: `0 12px 24px -6px ${colors.primary.glow}`,
  success: `0 8px 16px -4px ${colors.success.glow}`,
  error: `0 8px 16px -4px ${colors.error.glow}`,
  warning: `0 8px 16px -4px ${colors.warning.glow}`,
  
  // Ambient
  glass: '0 8px 32px 0 rgba(31, 38, 135, 0.07)',
};
