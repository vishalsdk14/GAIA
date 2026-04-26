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
import { Play, CheckCircle2, AlertCircle, Clock, ChevronRight, Activity } from 'lucide-react';
import { theme } from '@/theme';

const TaskList = ({ onSelect, selectedID }: { onSelect: (id: string) => void, selectedID?: string | null }) => {
  const { data: tasks, error, isLoading } = useSWR('/api/v1/tasks', fetcher, {
    refreshInterval: 5000,
  });

  if (isLoading) return (
    <div 
      className="text-center font-black animate-pulse"
      style={{ 
        padding: theme.spacing['2xl'],
        fontSize: theme.typography.size.sm, 
        letterSpacing: theme.typography.tracking.ultra, 
        color: theme.colors.text.muted 
      }}
    >
      RECONCILING...
    </div>
  );
  
  if (error) return (
    <div 
      className="text-center font-black uppercase tracking-widest"
      style={{ 
        padding: theme.spacing['2xl'],
        fontSize: theme.typography.size.sm, 
        color: theme.colors.error.DEFAULT 
      }}
    >
      Sync Error
    </div>
  );
  
  if (!tasks || tasks.length === 0) return (
    <div 
      className="text-center font-black uppercase tracking-widest"
      style={{ 
        padding: theme.spacing['2xl'],
        fontSize: theme.typography.size.sm, 
        color: theme.colors.text.dim 
      }}
    >
      No Active Missions
    </div>
  );

  const getStatusStyles = (status: string) => {
    switch (status) {
      case 'executing': return { color: theme.colors.primary.DEFAULT, bg: 'bg-blue-50', border: 'border-blue-100', icon: <Activity className="w-3.5 h-3.5 animate-pulse" /> };
      case 'completed': return { color: theme.colors.success.DEFAULT, bg: 'bg-emerald-50', border: 'border-emerald-100', icon: <CheckCircle2 className="w-3.5 h-3.5" /> };
      case 'failed': return { color: theme.colors.error.DEFAULT, bg: 'bg-red-50', border: 'border-red-100', icon: <AlertCircle className="w-3.5 h-3.5" /> };
      case 'planning': return { color: theme.colors.warning.DEFAULT, bg: 'bg-amber-50', border: 'border-amber-100', icon: <Clock className="w-3.5 h-3.5 animate-spin" /> };
      default: return { color: theme.colors.text.muted, bg: 'bg-slate-50', border: 'border-slate-100', icon: <ChevronRight className="w-3.5 h-3.5" /> };
    }
  };

  return (
    <div className="flex flex-col" style={{ gap: theme.spacing.md }}>
      {tasks.map((task: any) => {
        const isSelected = selectedID === task.task_id;
        const styles = getStatusStyles(task.status);
        
        return (
          <div 
            key={task.task_id}
            onClick={() => onSelect(task.task_id)}
            className={`
              relative border transition-all cursor-pointer group overflow-hidden
              ${isSelected 
                ? 'bg-white border-blue-200 shadow-xl shadow-blue-500/[0.08]' 
                : 'bg-white/40 border-slate-100 hover:bg-white hover:border-slate-200 hover:shadow-md'
              }
            `}
            style={{ 
              padding: theme.spacing.lg,
              borderRadius: theme.radius.xl 
            }}
          >
            {isSelected && (
              <div 
                className="absolute left-0 rounded-r-full bg-blue-500" 
                style={{ top: theme.spacing.sm, bottom: theme.spacing.sm, width: theme.spacing.xs }}
              />
            )}
            
            <div className="flex flex-col" style={{ gap: theme.spacing.md }}>
              <div className="flex items-start justify-between" style={{ gap: theme.spacing.md }}>
                <div className="flex-1 min-w-0">
                  <h3 
                    className={`font-bold leading-snug line-clamp-2 transition-colors ${isSelected ? 'text-slate-900' : 'text-slate-600'}`}
                    style={{ fontSize: theme.typography.size.base }}
                  >
                    {task.goal}
                  </h3>
                  <div className="flex items-center mt-2" style={{ gap: theme.spacing.sm }}>
                    <span 
                      className="font-bold uppercase tracking-tighter"
                      style={{ fontSize: theme.typography.size.tiny, color: theme.colors.text.muted }}
                    >
                      {task.task_id.split('-')[0]}
                    </span>
                    <span 
                      className="rounded-full bg-slate-200" 
                      style={{ width: theme.spacing.xs, height: theme.spacing.xs }}
                    />
                    <span 
                      className="font-bold uppercase tracking-tighter"
                      style={{ fontSize: theme.typography.size.tiny, color: theme.colors.text.muted }}
                    >
                      {formatDistanceToNow(new Date(task.created_at))} ago
                    </span>
                  </div>
                </div>
                <div 
                  className={`flex items-center justify-center border flex-shrink-0 transition-transform group-hover:scale-110 ${styles.bg} ${styles.border}`}
                  style={{ 
                    color: styles.color,
                    width: theme.spacing.iconButtonSm,
                    height: theme.spacing.iconButtonSm,
                    borderRadius: theme.radius.md
                  }}
                >
                  {styles.icon}
                </div>
              </div>

              <div className="flex items-center justify-between">
                 <div 
                    className={`px-3 py-1 rounded-full font-black uppercase tracking-widest border ${styles.bg} ${styles.border}`}
                    style={{ color: styles.color, fontSize: theme.typography.size.tiny }}
                 >
                    {task.status}
                 </div>
                 <div className="flex items-center opacity-20 group-hover:opacity-100 transition-opacity" style={{ gap: theme.spacing.xs }}>
                    <div className="rounded-full bg-slate-300" style={{ width: theme.spacing.xs, height: theme.spacing.xs }} />
                    <div className="rounded-full bg-slate-300" style={{ width: theme.spacing.xs, height: theme.spacing.xs }} />
                 </div>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};

export default TaskList;
