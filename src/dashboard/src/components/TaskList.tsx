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
import useSWR from 'swr';
import { fetcher } from '@/lib/api';
import { formatDistanceToNow } from 'date-fns';
import { Play, CheckCircle2, AlertCircle, Clock } from 'lucide-react';
import { theme } from '@/theme';

const TaskList = ({ onSelect, selectedID }: { onSelect: (id: string) => void, selectedID?: string | null }) => {
  const { data: tasks, error, isLoading } = useSWR('/api/v1/tasks', fetcher, {
    refreshInterval: 5000,
  });

  if (isLoading) return <div className="p-8 text-center" style={{ color: theme.colors.text.muted }}>Loading tasks...</div>;
  if (error) return <div className="p-8 text-center" style={{ color: theme.colors.error.DEFAULT }}>Failed to load tasks.</div>;
  if (!tasks || tasks.length === 0) return <div className="p-8 text-center" style={{ color: theme.colors.text.muted }}>No active tasks.</div>;

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'executing': return <Play className="w-4 h-4" style={{ color: theme.colors.primary.DEFAULT }} />;
      case 'completed': return <CheckCircle2 className="w-4 h-4" style={{ color: theme.colors.success.DEFAULT }} />;
      case 'failed': return <AlertCircle className="w-4 h-4" style={{ color: theme.colors.error.DEFAULT }} />;
      case 'planning': return <Clock className="w-4 h-4 animate-pulse" style={{ color: theme.colors.warning.DEFAULT }} />;
      default: return <Clock className="w-4 h-4" style={{ color: theme.colors.text.muted }} />;
    }
  };

  return (
    <div className="flex flex-col gap-4">
      {tasks.map((task: any) => {
        const isSelected = selectedID === task.task_id;
        return (
          <div 
            key={task.task_id}
            onClick={() => onSelect(task.task_id)}
            style={{ 
              backgroundColor: isSelected ? theme.colors.surface.elevated : theme.colors.glass,
              borderColor: isSelected ? theme.colors.primary.DEFAULT : theme.colors.border,
              boxShadow: isSelected ? theme.shadows.glow : 'none'
            }}
            className={`border transition-all cursor-pointer p-4 rounded-xl flex items-center justify-between group`}
          >
            <div className="flex items-center gap-4">
              <div className="p-2 rounded-lg bg-white/5 group-hover:bg-white/10 transition-colors">
                {getStatusIcon(task.status)}
              </div>
              <div>
                <h3 className="font-medium text-white/90 line-clamp-1">{task.goal}</h3>
                <p className="text-xs mt-1 uppercase tracking-tighter" style={{ color: theme.colors.text.secondary }}>
                  {task.task_id} • {formatDistanceToNow(new Date(task.created_at))} ago
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <div className="text-[10px] px-2 py-0.5 rounded-full border bg-white/5 uppercase font-bold tracking-widest" style={{ color: theme.colors.text.muted, borderColor: theme.colors.border }}>
                {task.status}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};

export default TaskList;
