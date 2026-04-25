'use client';

import React, { useState } from 'react';
import Header from '@/components/Header';
import TaskList from '@/components/TaskList';
import DAGView from '@/components/DAGView';
import useSWR from 'swr';
import { fetcher, approveStep, submitTask } from '@/lib/api';
import { Send, LayoutDashboard, Database, Activity, Shield } from 'lucide-react';
import { theme } from '@/theme';

export default function Dashboard() {
  const [selectedTaskID, setSelectedTaskID] = useState<string | null>(null);
  const [goal, setGoal] = useState('');

  const { data: selectedTask, mutate: mutateTask } = useSWR(
    selectedTaskID ? `/api/v1/tasks/${selectedTaskID}` : null,
    fetcher,
    { refreshInterval: 2000 }
  );

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!goal.trim()) return;
    try {
      const task = await submitTask(goal);
      setSelectedTaskID(task.task_id);
      setGoal('');
    } catch (err) {
      console.error(err);
    }
  };

  const handleApprove = async (stepID: string) => {
    if (!selectedTaskID) return;
    try {
      await approveStep(selectedTaskID, stepID);
      mutateTask();
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <main className="min-h-screen flex flex-col" style={{ backgroundColor: theme.colors.background, color: theme.colors.foreground }}>
      <Header />
      
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <aside className="border-r flex flex-col bg-black/20" style={{ width: theme.spacing.sidebarWidth, borderColor: theme.colors.border }}>
          <div className="p-4 border-b" style={{ borderColor: theme.colors.border }}>
            <form onSubmit={handleSubmit} className="relative">
              <input 
                type="text" 
                value={goal}
                onChange={(e) => setGoal(e.target.value)}
                placeholder="Submit new goal..."
                style={{ backgroundColor: theme.colors.surface.elevated, borderColor: theme.colors.border }}
                className="w-full border rounded-xl py-2 pl-4 pr-10 text-sm focus:outline-none focus:border-indigo-500/50 transition-all placeholder:text-white/20"
              />
              <button type="submit" className="absolute right-2 top-1/2 -translate-y-1/2 transition-colors" style={{ color: theme.colors.text.secondary }}>
                <Send className="w-4 h-4" />
              </button>
            </form>
          </div>
          
          <div className="flex-1 overflow-y-auto p-4 custom-scrollbar">
            <h2 className="text-[10px] font-bold uppercase tracking-[0.2em] mb-4 px-2" style={{ color: theme.colors.text.muted }}>Active Tasks</h2>
            <TaskList onSelect={setSelectedTaskID} selectedID={selectedTaskID} />
          </div>
          
          <div className="p-4 border-t bg-black/40" style={{ borderColor: theme.colors.border }}>
            <nav className="flex flex-col gap-1">
              <a href="#" className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium" style={{ backgroundColor: theme.colors.primary.subtle, color: theme.colors.primary.DEFAULT }}>
                <LayoutDashboard className="w-4 h-4" /> Dashboard
              </a>
              <a href="#" className="flex items-center gap-3 px-3 py-2 rounded-lg transition-all text-sm font-medium hover:bg-white/5" style={{ color: theme.colors.text.secondary }}>
                <Database className="w-4 h-4" /> State Registry
              </a>
              <a href="#" className="flex items-center gap-3 px-3 py-2 rounded-lg transition-all text-sm font-medium hover:bg-white/5" style={{ color: theme.colors.text.secondary }}>
                <Activity className="w-4 h-4" /> Agent Health
              </a>
              <a href="#" className="flex items-center gap-3 px-3 py-2 rounded-lg transition-all text-sm font-medium hover:bg-white/5" style={{ color: theme.colors.text.secondary }}>
                <Shield className="w-4 h-4" /> Policy Logs
              </a>
            </nav>
          </div>
        </aside>

        {/* Main Content */}
        <section className="flex-1 flex flex-col bg-black/40 p-6 relative">
          {!selectedTask ? (
            <div className="flex-1 flex flex-col items-center justify-center opacity-20 pointer-events-none">
              <div className="w-24 h-24 border-2 border-dashed rounded-full flex items-center justify-center mb-6" style={{ borderColor: theme.colors.border }}>
                <LayoutDashboard className="w-10 h-10" />
              </div>
              <h2 className="text-2xl font-light tracking-widest uppercase">Select a Task</h2>
              <p className="text-sm mt-2">Monitor real-time agent coordination</p>
            </div>
          ) : (
            <div className="flex-1 flex flex-col gap-6">
              <div className="flex items-center justify-between">
                <div>
                  <div className="flex items-center gap-3">
                    <h2 className="text-2xl font-bold" style={{ color: theme.colors.text.primary }}>{selectedTask.goal}</h2>
                    <span className="px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-tighter border" style={{ backgroundColor: theme.colors.primary.subtle, color: theme.colors.primary.DEFAULT, borderColor: theme.colors.primary.subtle }}>
                      {selectedTask.status}
                    </span>
                  </div>
                  <p className="text-xs mt-1 uppercase font-mono" style={{ color: theme.colors.text.muted }}>{selectedTask.task_id}</p>
                </div>
                <div className="flex items-center gap-4">
                  <div className="text-right">
                    <p className="text-[10px] uppercase font-bold tracking-widest" style={{ color: theme.colors.text.muted }}>Execution Progress</p>
                    <div className="w-32 h-1.5 rounded-full mt-1 overflow-hidden border" style={{ backgroundColor: theme.colors.surface.base, borderColor: theme.colors.border }}>
                      <div 
                        className="h-full transition-all duration-1000"
                        style={{ 
                          width: `${(selectedTask.plan.filter((s:any) => s.status === 'done').length / selectedTask.plan.length) * 100}%`,
                          background: `linear-gradient(to right, ${theme.colors.primary.DEFAULT}, #a855f7)`
                        }}
                      />
                    </div>
                  </div>
                </div>
              </div>

              <div className="flex-1 min-h-0 relative">
                <DAGView 
                  plan={selectedTask.plan} 
                  taskID={selectedTask.task_id} 
                  onApprove={handleApprove} 
                />
              </div>

              {/* Console/Logs */}
              <div className="h-48 glass rounded-2xl p-4 overflow-hidden flex flex-col">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-[10px] font-black uppercase tracking-widest" style={{ color: theme.colors.text.muted }}>Event Stream</h3>
                  <div className="w-2 h-2 rounded-full animate-pulse shadow-[0_0_8px_rgba(99,102,241,0.8)]" style={{ backgroundColor: theme.colors.primary.DEFAULT }} />
                </div>
                <div className="flex-1 font-mono text-[11px] overflow-y-auto custom-scrollbar flex flex-col-reverse gap-1" style={{ color: theme.colors.text.secondary }}>
                  {selectedTask.plan.filter((s:any) => s.status === 'done').map((s:any) => (
                    <div key={s.step_id} className="flex gap-2">
                      <span style={{ color: theme.colors.success.DEFAULT }}>✓</span>
                      <span>Completed <span className="font-bold text-white/80">{s.capability}</span> in {s.step_id}</span>
                    </div>
                  ))}
                  <div className="flex gap-2" style={{ color: theme.colors.primary.DEFAULT }}>
                    <span className="animate-pulse">&gt;</span>
                    <span>Monitoring control loop heartbeat...</span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </section>
      </div>
    </main>
  );
}
