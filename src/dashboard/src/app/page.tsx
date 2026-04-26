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

import React, { useState } from 'react';
import Header from '@/components/Header';
import Sidebar from '@/components/Sidebar';
import DAGView from '@/components/DAGView';
import NewMissionModal from '@/components/NewMissionModal';
import Button from '@/components/Button';
import useSWR from 'swr';
import { fetcher, approveStep, submitTask } from '@/lib/api';
import { 
  Terminal, 
  Cpu, 
  Activity, 
  ShieldAlert, 
  Layers, 
  Clock,
  ExternalLink,
  ChevronRight,
  Info
} from 'lucide-react';
import { theme } from '@/theme';

/**
 * Dashboard is the primary entry point for the GAIA Control Center.
 */
export default function Dashboard() {
  const [selectedTaskID, setSelectedTaskID] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const { data: selectedTask, mutate: mutateTask } = useSWR(
    selectedTaskID ? `/api/v1/tasks/${selectedTaskID}` : null,
    fetcher,
    { refreshInterval: 2000 }
  );

  const handleCreateMission = async (goal: string) => {
    try {
      const task = await submitTask(goal);
      setSelectedTaskID(task.task_id);
    } catch (err) {
      console.error('Failed to initialize mission:', err);
    }
  };

  const handleApprove = async (stepID: string) => {
    if (!selectedTaskID) return;
    try {
      await approveStep(selectedTaskID, stepID);
      mutateTask();
    } catch (err) {
      console.error('Step approval failed:', err);
    }
  };

  return (
    <main className="h-screen flex flex-col relative overflow-hidden" style={{ color: theme.colors.text.primary, backgroundColor: theme.colors.background }}>
      <div className="grid-bg absolute inset-0 pointer-events-none opacity-40" />
      
      <Header />
      
      <div className="flex-1 flex overflow-hidden">
        {/* Navigation Rail */}
        <aside 
          className="flex-shrink-0 border-r flex flex-col items-center z-10 bg-white"
          style={{ 
            width: theme.spacing.railWidth, 
            borderColor: theme.colors.border.subtle,
            paddingTop: theme.spacing.xl,
            gap: theme.spacing.md
          }}
        >
           {[
             { Icon: Layers, active: true },
             { Icon: Cpu, active: false },
             { Icon: Activity, active: false },
             { Icon: ShieldAlert, active: false }
           ].map((item, idx) => (
             <Button 
               key={idx}
               variant={item.active ? 'primary' : 'ghost'}
               size="icon"
               icon={item.Icon}
               style={{ 
                 backgroundColor: item.active ? theme.colors.primary.DEFAULT : 'transparent',
                 color: item.active ? '#ffffff' : theme.colors.text.muted,
                 borderColor: 'transparent'
               }}
             />
           ))}
        </aside>

        {/* Modular Sidebar */}
        <Sidebar 
          onSelectTask={setSelectedTaskID} 
          selectedTaskID={selectedTaskID} 
          onOpenNewMission={() => setIsModalOpen(true)}
        />

        {/* Primary Stage */}
        <section className="flex-1 flex flex-col relative bg-white/40 min-w-0">
          {!selectedTask ? (
            <div className="flex-1 flex flex-col items-center justify-center relative">
              <div 
                className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full blur-[140px] pointer-events-none bg-blue-500/[0.04]" 
                style={{ width: '600px', height: '600px' }}
              />
              <div className="relative z-10 flex flex-col items-center">
                <div 
                  className="border flex items-center justify-center bg-white shadow-xl"
                  style={{ 
                    width: '6.5rem', 
                    height: '6.5rem', 
                    borderRadius: theme.radius.xl,
                    borderColor: theme.colors.border.subtle 
                  }}
                >
                   <Terminal className="w-10 h-10 text-blue-500/40" strokeWidth={1.5} />
                </div>
                <h2 
                  className="font-black italic mb-4"
                  style={{ 
                    fontSize: theme.typography.size['4xl'],
                    letterSpacing: theme.typography.tracking.tighter,
                    color: theme.colors.text.primary
                  }}
                >
                  Awaiting Goal
                </h2>
                <p 
                  className="font-bold uppercase"
                  style={{ 
                    fontSize: theme.typography.size.sm, 
                    letterSpacing: theme.typography.tracking.widest,
                    color: theme.colors.text.muted 
                  }}
                >
                  Select or Initialize a mission to begin
                </p>
                
                <Button 
                  variant="primary" 
                  size="lg" 
                  onClick={() => setIsModalOpen(true)}
                  className="mt-10"
                  style={{ paddingLeft: theme.spacing['2xl'], paddingRight: theme.spacing['2xl'] }}
                >
                  Deploy New Mission
                </Button>
              </div>
            </div>
          ) : (
            <div className="flex-1 flex flex-col overflow-hidden">
              {/* Context Bar */}
              <div 
                className="flex items-center justify-between border-b bg-white/80 backdrop-blur-md" 
                style={{ 
                  borderColor: theme.colors.border.subtle, 
                  height: theme.spacing['4xl'],
                  paddingLeft: theme.spacing.xl,
                  paddingRight: theme.spacing.xl
                }}
              >
                <div className="flex items-center gap-6">
                   <div className="flex flex-col">
                      <div className="flex items-center gap-4">
                        <h2 
                          className="font-black"
                          style={{ 
                            fontSize: theme.typography.size['2xl'],
                            letterSpacing: theme.typography.tracking.tight,
                            color: theme.colors.text.primary
                          }}
                        >
                          {selectedTask.goal}
                        </h2>
                        <div 
                          className="px-3 py-1 font-black uppercase border shadow-sm" 
                          style={{ 
                            fontSize: theme.typography.size.xs,
                            letterSpacing: theme.typography.tracking.widest,
                            backgroundColor: theme.colors.primary.subtle, 
                            color: theme.colors.primary.DEFAULT, 
                            borderColor: theme.colors.primary.glow,
                            borderRadius: theme.radius.full
                          }}
                        >
                          {selectedTask.status}
                        </div>
                      </div>
                      <div className="flex items-center gap-5 mt-1.5">
                         <span 
                           className="font-bold flex items-center gap-1.5"
                           style={{ fontSize: theme.typography.size.sm, color: theme.colors.text.muted }}
                         >
                            <Info className="w-3.5 h-3.5" /> {selectedTask.task_id}
                         </span>
                         <span 
                           className="font-bold flex items-center gap-1.5"
                           style={{ fontSize: theme.typography.size.sm, color: theme.colors.text.muted }}
                         >
                            <Clock className="w-3.5 h-3.5" /> Started 2m ago
                         </span>
                      </div>
                   </div>
                </div>
                
                <div className="flex items-center gap-10">
                   <div className="flex flex-col items-end">
                      <span 
                        className="font-black uppercase mb-1.5"
                        style={{ 
                          fontSize: theme.typography.size.xs, 
                          letterSpacing: theme.typography.tracking.widest,
                          color: theme.colors.text.muted 
                        }}
                      >
                        Execution Index
                      </span>
                      <div className="flex items-center gap-4">
                         <div className="w-48 h-2 rounded-full bg-slate-100 overflow-hidden shadow-inner">
                            <div 
                              className="h-full transition-all duration-1000 rounded-full"
                              style={{ 
                                width: `${(selectedTask.plan.filter((s:any) => s.status === 'done').length / selectedTask.plan.length) * 100}%`,
                                background: `linear-gradient(to right, ${theme.colors.primary.DEFAULT}, ${theme.colors.secondary.DEFAULT})`,
                                boxShadow: theme.shadows.primary
                              }}
                            />
                         </div>
                         <span className="font-black text-blue-600" style={{ fontSize: theme.typography.size.md }}>
                           {Math.round((selectedTask.plan.filter((s:any) => s.status === 'done').length / selectedTask.plan.length) * 100)}%
                         </span>
                      </div>
                   </div>
                   <Button size="icon" variant="secondary" icon={ExternalLink} />
                </div>
              </div>

              {/* DAG Canvas */}
              <div className="flex-1 relative min-h-0 bg-slate-50/50">
                <DAGView 
                  plan={selectedTask.plan} 
                  taskID={selectedTask.task_id} 
                  onApprove={handleApprove} 
                />
              </div>

              {/* Event Feed */}
              <div 
                className="border-t flex flex-col bg-white/90 backdrop-blur-xl" 
                style={{ 
                  borderColor: theme.colors.border.subtle,
                  height: theme.spacing.feedHeight 
                }}
              >
                 <div 
                   className="flex items-center justify-between border-b" 
                   style={{ borderColor: theme.colors.border.subtle, paddingLeft: theme.spacing.xl, paddingRight: theme.spacing.xl, paddingTop: theme.spacing.md, paddingBottom: theme.spacing.md }}
                 >
                    <div className="flex items-center gap-2.5">
                       <Terminal className="w-4 h-4 text-slate-400" strokeWidth={2} />
                       <span 
                         className="font-black uppercase"
                         style={{ 
                           fontSize: theme.typography.size.sm, 
                           letterSpacing: theme.typography.tracking.widest,
                           color: theme.colors.text.muted 
                         }}
                       >
                         Orchestration Feed
                       </span>
                    </div>
                    <div className="flex items-center gap-2.5">
                       <span 
                         className="font-bold uppercase"
                         style={{ fontSize: theme.typography.size.xs, letterSpacing: theme.typography.tracking.widest, color: theme.colors.text.dim }}
                       >
                         v1.2 Sync
                       </span>
                       <div className="w-2 h-2 rounded-full bg-blue-500 shadow-lg" />
                    </div>
                 </div>
                 <div className="flex-1 overflow-y-auto p-8 font-mono text-[12px] premium-scrollbar flex flex-col-reverse gap-4">
                    {selectedTask.plan.filter((s:any) => s.status === 'done').map((s:any) => (
                      <div key={s.step_id} className="flex gap-6 items-start group">
                        <span className="text-emerald-500 font-bold flex-shrink-0">✓ OK</span>
                        <div className="flex-1">
                           <div className="text-slate-700 font-medium leading-relaxed">Agent completed <span className="text-blue-600 font-bold">{s.capability}</span> successfully.</div>
                           <div className="text-[10px] text-slate-400 mt-1.5 flex items-center gap-3">
                              <span className="px-2 py-0.5 bg-slate-100 rounded">ID: {s.step_id}</span>
                              <ChevronRight className="w-3 h-3" />
                              <span className="px-2 py-0.5 bg-slate-100 rounded">CHANNEL_VERIFIED</span>
                           </div>
                        </div>
                        <span className="text-slate-300 font-medium tabular-nums opacity-0 group-hover:opacity-100 transition-opacity">12:05:44</span>
                      </div>
                    ))}
                    <div className="flex gap-6 items-center">
                       <span className="text-blue-500 font-bold flex-shrink-0 animate-pulse">● SYNC</span>
                       <span className="text-slate-400 font-medium italic">Monitoring kernel...</span>
                    </div>
                 </div>
              </div>
            </div>
          )}
        </section>
      </div>

      <NewMissionModal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
        onSubmit={handleCreateMission} 
      />
    </main>
  );
}
