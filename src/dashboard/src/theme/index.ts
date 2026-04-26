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
import { shadows } from './shadows';

/**
 * theme is the unified design system object used by all dashboard components.
 */
export const theme = {
  colors,
  spacing,
  typography,
  shadows,
  radius: spacing.borderRadius,
  strokeWidth: {
    thin: 1,
    base: 2,
    bold: 2.5,
    heavy: 3,
  }
};

export * from './colors';
export * from './spacing';
export * from './typography';
export * from './shadows';
