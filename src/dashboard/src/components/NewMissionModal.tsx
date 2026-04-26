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
import { theme } from '@/theme';
import { Send, Terminal, ShieldCheck } from 'lucide-react';
import Modal from './Modal';
import Button from './Button';

interface NewMissionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (goal: string) => void;
}

/**
 * NewMissionModal provides a specialized interface for mission initialization.
 * 100% theme-driven: Zero magic numbers across spacing, typography, and colors.
 */
const NewMissionModal = ({ isOpen, onClose, onSubmit }: NewMissionModalProps) => {
  const [goal, setGoal] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!goal.trim()) return;
    onSubmit(goal);
    setGoal('');
    onClose();
  };

  return (
    <Modal 
      isOpen={isOpen} 
      onClose={onClose} 
      title="Initialize Mission" 
      subtitle="GAIA ORCHESTRATOR v1.2"
    >
      <form onSubmit={handleSubmit} style={{ gap: theme.spacing.xl }} className="flex flex-col">
        <div style={{ gap: theme.spacing.sm }} className="flex flex-col">
          <label 
            className="font-black uppercase ml-1"
            style={{ 
              fontSize: theme.typography.size.sm, 
              letterSpacing: theme.typography.tracking.widest,
              color: theme.colors.text.muted 
            }}
          >
            Strategic Objective
          </label>
          <div className="relative group">
            <div 
              className="absolute"
              style={{ 
                top: theme.spacing.md, 
                left: theme.spacing.lg,
                color: theme.colors.text.dim 
              }}
            >
              <Terminal className="w-5 h-5" />
            </div>
            <textarea 
              autoFocus
              value={goal}
              onChange={(e) => setGoal(e.target.value)}
              placeholder="Describe the mission objective in natural language..."
              className="w-full border focus:outline-none focus:border-blue-400 transition-all"
              style={{ 
                padding: theme.spacing.lg,
                paddingLeft: theme.spacing['2xl'],
                minHeight: theme.spacing['4xl'],
                fontSize: theme.typography.size.base,
                borderRadius: theme.radius.xl,
                lineHeight: theme.typography.lineHeight.relaxed,
                backgroundColor: theme.colors.surface.low,
                borderColor: theme.colors.border.subtle,
                color: theme.colors.text.primary
              }}
            />
          </div>
        </div>

        <div 
          className="flex items-center justify-between border-t"
          style={{ borderColor: theme.colors.border.subtle, paddingTop: theme.spacing.lg }}
        >
           <div 
             className="flex items-center font-bold uppercase"
             style={{ 
               gap: theme.spacing.sm,
               fontSize: theme.typography.size.xs,
               letterSpacing: theme.typography.tracking.widest,
               color: theme.colors.text.muted
             }}
           >
              <ShieldCheck 
                className="w-4 h-4" 
                style={{ color: theme.colors.success.DEFAULT }} 
              />
              Kernel Validated
           </div>
           <div className="flex items-center" style={{ gap: theme.spacing.sm }}>
             <Button variant="ghost" onClick={onClose}>
               Cancel
             </Button>
             <Button variant="primary" type="submit" icon={Send}>
               Deploy Mission
             </Button>
           </div>
        </div>
      </form>
    </Modal>
  );
};

export default NewMissionModal;
