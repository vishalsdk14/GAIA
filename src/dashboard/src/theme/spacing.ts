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

// src/theme/spacing.ts

/**
 * spacing defines the layout and dimensional scale for the GAIA dashboard.
 */
export const spacing = {
  headerHeight: '4.5rem',
  
  // Layout Columns
  railWidth: '5.5rem',
  sidebarWidth: '22rem',
  detailsWidth: '24rem',
  
  // Component Dimensions
  iconButtonSm: '2rem',
  iconButtonMd: '2.5rem',
  iconButtonLg: '3rem',
  logoIconSize: '2.4rem',
  dotIndicatorSize: '0.6rem',
  
  // Feed/Log Height
  feedHeight: '18rem',
  
  containerMax: '96rem',
  
  // Padding & Margin Scale
  xs: '0.25rem',
  sm: '0.5rem',
  md: '1rem',
  lg: '1.5rem',
  xl: '2rem',
  '2xl': '3rem',
  '3xl': '4rem',
  '4xl': '6rem',
  
  // Component specific
  nodeMinWidth: '220px',
  nodeVerticalGap: 220, // Numeric for ReactFlow math
  nodeHorizontalOffset: 400, // Numeric for ReactFlow math
  nodeTopOffset: 100, // Numeric for ReactFlow math
  flowGridGap: 40,
  
  borderRadius: {
    xs: '0.25rem',
    sm: '0.5rem',
    md: '0.75rem',
    lg: '1rem',
    xl: '1.25rem',
    '2xl': '1.5rem',
    full: '9999px',
  }
};
