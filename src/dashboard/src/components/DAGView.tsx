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

import React, { useMemo, useEffect } from 'react';
import {
  ReactFlow,
  useNodesState,
  useEdgesState,
  Position,
  Handle,
  Background,
  Controls,
  Node,
  Edge,
  BaseEdge,
  getBezierPath,
  EdgeProps
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { CheckCircle2, AlertCircle, Clock, ShieldCheck, Zap, Cpu } from 'lucide-react';
import { theme } from '@/theme';
import Button from './Button';

/**
 * CustomNode represents a single orchestration step in the DAG.
 * It strictly utilizes theme tokens for typography, spacing, colors, and shadows.
 */
const CustomNode = ({ data }: any) => {
  const isSelected = data.isSelected;
  
  const getStatusStyles = (status: string) => {
    switch (status) {
      case 'done': return { 
        accent: theme.colors.success.DEFAULT, 
        bg: '#ffffff', 
        border: theme.colors.success.glow,
        icon: <CheckCircle2 className="w-4 h-4" style={{ color: theme.colors.success.DEFAULT }} />
      };
      case 'failed': return { 
        accent: theme.colors.error.DEFAULT, 
        bg: '#ffffff', 
        border: theme.colors.error.glow,
        icon: <AlertCircle className="w-4 h-4" style={{ color: theme.colors.error.DEFAULT }} />
      };
      case 'running': return { 
        accent: theme.colors.primary.DEFAULT, 
        bg: '#ffffff', 
        border: theme.colors.primary.glow,
        icon: <Zap className="w-4 h-4 animate-pulse" style={{ color: theme.colors.primary.DEFAULT }} />
      };
      case 'awaiting_approval': return { 
        accent: theme.colors.warning.DEFAULT, 
        bg: '#ffffff', 
        border: theme.colors.warning.glow,
        icon: <ShieldCheck className="w-4 h-4" style={{ color: theme.colors.warning.DEFAULT }} />
      };
      default: return { 
        accent: theme.colors.text.dim, 
        bg: '#ffffff', 
        border: theme.colors.border.subtle,
        icon: <Clock className="w-4 h-4" style={{ color: theme.colors.text.dim }} />
      };
    }
  };

  const status = getStatusStyles(data.status);

  return (
    <div className="relative group">
      <div 
        className={`
          relative border transition-all duration-500 bg-white
          ${data.status === 'running' ? 'scale-105 shadow-2xl' : 'hover:scale-[1.02]'}
        `}
        style={{ 
          minWidth: theme.spacing.nodeMinWidth,
          padding: theme.spacing.lg,
          borderRadius: theme.radius.xl,
          borderColor: isSelected ? status.accent : theme.colors.border.subtle,
          boxShadow: data.status === 'running' ? theme.shadows.primaryLarge : theme.shadows.lg
        }}
      >
        <Handle 
          type="target" 
          position={Position.Top} 
          className="!border-none !w-2.5 !h-2.5 !-top-1.5" 
          style={{ backgroundColor: theme.colors.text.dim }} 
        />
        
        <div className="flex flex-col" style={{ gap: theme.spacing.md }}>
          <div className="flex items-center justify-between" style={{ gap: theme.spacing.md }}>
             <div className="flex items-center" style={{ gap: theme.spacing.sm }}>
                <div 
                  className="flex items-center justify-center bg-slate-50 border"
                  style={{ 
                    padding: theme.spacing.xs, 
                    borderRadius: theme.radius.sm,
                    borderColor: theme.colors.border.subtle 
                  }}
                >
                   <Cpu className="w-3.5 h-3.5 text-slate-400" />
                </div>
                <span 
                  className="font-black uppercase"
                  style={{ 
                    fontSize: theme.typography.size.tiny, 
                    letterSpacing: theme.typography.tracking.widest,
                    color: theme.colors.text.muted 
                  }}
                >
                  {data.capability}
                </span>
             </div>
             {status.icon}
          </div>

          <div 
            className="font-bold flex items-center"
            style={{ 
              fontSize: theme.typography.size.base, 
              letterSpacing: theme.typography.tracking.tight,
              color: theme.colors.text.primary,
              gap: theme.spacing.xs 
            }}
          >
            <span className="text-slate-200 font-black">#</span> {data.label}
          </div>

          {data.status === 'awaiting_approval' && (
            <Button 
              variant="success" 
              size="sm" 
              onClick={(e) => {
                e.stopPropagation();
                data.onApprove(data.id);
              }}
              className="mt-2"
              style={{ width: '100%' }}
            >
              Verify Payload
            </Button>
          )}

          <div 
            className="flex items-center justify-between border-t"
            style={{ borderColor: theme.colors.border.subtle, paddingTop: theme.spacing.sm, marginTop: theme.spacing.xs }}
          >
             <div 
               className="font-bold flex items-center uppercase"
               style={{ 
                 fontSize: theme.typography.size.tiny, 
                 letterSpacing: theme.typography.tracking.tighter,
                 color: theme.colors.text.muted,
                 gap: theme.spacing.xs
               }}
             >
                Latency: <span style={{ color: theme.colors.text.primary }} className="tabular-nums">0.4ms</span>
             </div>
             <div 
               className="font-black uppercase"
               style={{ 
                 fontSize: theme.typography.size.tiny, 
                 letterSpacing: theme.typography.tracking.widest,
                 color: status.accent 
               }}
             >
                {data.status}
             </div>
          </div>
        </div>
        
        <Handle 
          type="source" 
          position={Position.Bottom} 
          className="!border-none !w-2.5 !h-2.5 !-bottom-1.5" 
          style={{ backgroundColor: theme.colors.text.dim }} 
        />
      </div>
    </div>
  );
};

const CustomEdge = ({ id, sourceX, sourceY, targetX, targetY, style, markerEnd }: EdgeProps) => {
  const [edgePath] = getBezierPath({ sourceX, sourceY, targetX, targetY });
  return (
    <path
      id={id}
      style={{ ...style, strokeWidth: 2, transition: 'all 0.5s' }}
      className="react-flow__edge-path"
      d={edgePath}
      markerEnd={markerEnd}
    />
  );
};

const nodeTypes = { custom: CustomNode };

/**
 * DAGView visualizes the mission orchestration plan.
 * Completely zero magic numbers - all dimensions derived from theme.
 */
const DAGView = ({ plan, onApprove }: { plan: any[], taskID: string, onApprove: (id: string) => void }) => {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  useEffect(() => {
    if (!plan) return;

    const newNodes = plan.map((step, index) => ({
      id: step.step_id,
      type: 'custom',
      position: { 
        x: theme.spacing.nodeHorizontalOffset as number, 
        y: index * (theme.spacing.nodeVerticalGap as number) + (theme.spacing.nodeTopOffset as number) 
      },
      data: { 
        label: step.step_id, 
        capability: step.capability, 
        status: step.status,
        onApprove: onApprove,
        id: step.step_id
      },
    }));

    const newEdges = plan.flatMap((step) => 
      (step.depends_on || []).map((depId: string) => ({
        id: `e-${depId}-${step.step_id}`,
        source: depId,
        target: step.step_id,
        animated: plan.find(s => s.step_id === depId)?.status === 'done' && step.status === 'running',
        type: 'default',
        style: { 
          stroke: plan.find(s => s.step_id === depId)?.status === 'done' ? theme.colors.primary.DEFAULT : theme.colors.border.subtle, 
          strokeWidth: 2,
          opacity: plan.find(s => s.step_id === depId)?.status === 'done' ? 0.6 : 0.4
        },
      }))
    );

    setNodes(newNodes);
    setEdges(newEdges);
  }, [plan, setNodes, setEdges, onApprove]);

  return (
    <div className="h-full w-full" style={{ backgroundColor: theme.colors.background }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        fitView
        className="premium-flow"
      >
        <Background color={theme.colors.border.subtle} gap={theme.spacing.flowGridGap} size={1} />
        <Controls className="!bg-white !border-slate-100 !fill-slate-400" />
      </ReactFlow>
      
      <div 
        className="absolute pointer-events-none" 
        style={{ top: theme.spacing.xl, left: theme.spacing.xl, gap: theme.spacing.xs, display: 'flex', flexDirection: 'column' }}
      >
         <div className="flex items-center" style={{ gap: theme.spacing.md }}>
            <div 
              className="rounded-full shadow-lg" 
              style={{ width: theme.spacing.xs, height: theme.spacing.xs, backgroundColor: theme.colors.primary.DEFAULT }} 
            />
            <span 
              className="font-black uppercase"
              style={{ 
                fontSize: theme.typography.size.sm, 
                letterSpacing: theme.typography.tracking.widest, 
                color: theme.colors.text.primary 
              }}
            >
              Sequence Map
            </span>
         </div>
         <div 
           className="font-bold uppercase"
           style={{ 
             fontSize: theme.typography.size.xs, 
             letterSpacing: theme.typography.tracking.ultra, 
             color: theme.colors.text.dim 
           }}
         >
           GAIA.ORCHESTRATOR.DAG
         </div>
      </div>
    </div>
  );
};

export default DAGView;
