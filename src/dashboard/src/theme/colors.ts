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

import { theme } from './index';

/**
 * colors defines the semantic color palette for the GAIA dashboard.
 * It maps raw hex/rgba values to semantic tokens like 'primary', 'success', etc.
 */
export const colors = {
  background: '#050505',
  foreground: '#ededed',
  glass: 'rgba(255, 255, 255, 0.05)',
  border: 'rgba(255, 255, 255, 0.1)',
  
  primary: {
    DEFAULT: '#6366f1',
    glow: 'rgba(99, 102, 241, 0.5)',
    subtle: 'rgba(99, 102, 241, 0.1)',
  },
  success: {
    DEFAULT: '#22c55e',
    glow: 'rgba(34, 197, 94, 0.5)',
    subtle: 'rgba(34, 197, 94, 0.1)',
  },
  warning: {
    DEFAULT: '#eab308',
    glow: 'rgba(234, 179, 8, 0.5)',
    subtle: 'rgba(234, 179, 8, 0.1)',
  },
  error: {
    DEFAULT: '#ef4444',
    glow: 'rgba(239, 68, 68, 0.5)',
    subtle: 'rgba(239, 68, 68, 0.1)',
  },
  
  text: {
    primary: 'rgba(255, 255, 255, 0.9)',
    secondary: 'rgba(255, 255, 255, 0.4)',
    muted: 'rgba(255, 255, 255, 0.2)',
  },
  
  surface: {
    base: 'rgba(0, 0, 0, 0.2)',
    elevated: 'rgba(255, 255, 255, 0.05)',
    overlay: 'rgba(255, 255, 255, 0.1)',
  }
};
