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

/**
 * colors defines the semantic color palette for the GAIA dashboard.
 * Focused on a premium Light Mode ("Aura Light") aesthetic.
 */
export const colors = {
  // Pure & Clean Light Palette
  background: '#ffffff',
  foreground: '#0f172a',
  
  // Glassmorphism tokens (Frosted White)
  glass: {
    thin: 'rgba(255, 255, 255, 0.4)',
    base: 'rgba(255, 255, 255, 0.6)',
    thick: 'rgba(255, 255, 255, 0.8)',
  },
  
  border: {
    subtle: 'rgba(0, 0, 0, 0.05)',
    base: 'rgba(0, 0, 0, 0.08)',
    strong: 'rgba(0, 0, 0, 0.12)',
  },
  
  // Refined Accents
  primary: {
    DEFAULT: '#2563eb', // Clean Blue
    glow: 'rgba(37, 99, 235, 0.15)',
    subtle: 'rgba(37, 99, 235, 0.05)',
    deep: 'rgba(37, 99, 235, 0.02)',
  },
  secondary: {
    DEFAULT: '#4f46e5', // Indigo
    glow: 'rgba(79, 70, 229, 0.15)',
    subtle: 'rgba(79, 70, 229, 0.05)',
    deep: 'rgba(79, 70, 229, 0.02)',
  },
  success: {
    DEFAULT: '#059669', // Emerald
    glow: 'rgba(5, 150, 105, 0.15)',
    subtle: 'rgba(5, 150, 105, 0.05)',
    deep: 'rgba(5, 150, 105, 0.02)',
  },
  warning: {
    DEFAULT: '#d97706', // Amber
    glow: 'rgba(217, 119, 6, 0.15)',
    subtle: 'rgba(217, 119, 6, 0.05)',
    deep: 'rgba(217, 119, 6, 0.02)',
  },
  error: {
    DEFAULT: '#dc2626', // Red
    glow: 'rgba(220, 38, 38, 0.15)',
    subtle: 'rgba(220, 38, 38, 0.05)',
    deep: 'rgba(220, 38, 38, 0.02)',
  },
  
  text: {
    primary: '#0f172a',
    secondary: '#475569',
    muted: '#94a3b8',
    dim: '#cbd5e1',
  },
  
  surface: {
    base: '#ffffff',
    low: '#f8fafc',
    mid: '#f1f5f9',
    high: '#e2e8f0',
    overlay: 'rgba(0, 0, 0, 0.02)',
    deep: '#f8fafc',
  }
};
