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
import TaskList from './TaskList';
import { Plus } from 'lucide-react';
import { theme } from '@/theme';
import Button from './Button';

interface SidebarProps {
  onSelectTask: (id: string) => void;
  selectedTaskID: string | null;
  onOpenNewMission: () => void;
}

/**
 * Sidebar component provides the mission browsing interface.
 * 100% theme-driven: Zero magic numbers across spacing and typography.
 */
const Sidebar = ({ onSelectTask, selectedTaskID, onOpenNewMission }: SidebarProps) => {
  return (
    <aside 
      className="flex-shrink-0 border-r flex flex-col backdrop-blur-sm z-10"
      style={{ 
        width: theme.spacing.sidebarWidth,
        borderColor: theme.colors.border.subtle,
        backgroundColor: theme.colors.glass.thin
      }}
    >
      <div style={{ padding: theme.spacing.lg }}>
        <div 
          className="flex items-center justify-between" 
          style={{ marginBottom: theme.spacing['2xl'] }}
        >
          <h2 
            className="font-black uppercase"
            style={{ 
              fontSize: theme.typography.size.sm, 
              letterSpacing: theme.typography.tracking.widest,
              color: theme.colors.text.muted 
            }}
          >
            Active Missions
          </h2>
          <Button 
            variant="secondary" 
            size="icon" 
            icon={Plus} 
            onClick={onOpenNewMission} 
          />
        </div>
      </div>
      
      <div 
        className="flex-1 overflow-y-auto premium-scrollbar"
        style={{ 
          paddingLeft: theme.spacing.lg, 
          paddingRight: theme.spacing.lg, 
          paddingBottom: theme.spacing['2xl'] 
        }}
      >
        <TaskList onSelect={onSelectTask} selectedID={selectedTaskID} />
      </div>
    </aside>
  );
};

export default Sidebar;
