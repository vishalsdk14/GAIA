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
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Play, CheckCircle2, AlertCircle, Clock, ShieldCheck } from 'lucide-react';
import { theme } from '@/theme';

const CustomNode = ({ data }: any) => {
  const getStatusStyles = (status: string) => {
    switch (status) {
      case 'done': return { 
        borderColor: 'rgba(34, 197, 94, 0.5)', 
        backgroundColor: theme.colors.success.subtle, 
        color: theme.colors.success.DEFAULT 
      };
      case 'failed': return { 
        borderColor: 'rgba(239, 68, 68, 0.5)', 
        backgroundColor: theme.colors.error.subtle, 
        color: theme.colors.error.DEFAULT 
      };
      case 'running': return { 
        borderColor: 'rgba(99, 102, 241, 0.5)', 
        backgroundColor: theme.colors.primary.subtle, 
        color: theme.colors.primary.DEFAULT 
      };
      case 'awaiting_approval': return { 
        borderColor: 'rgba(234, 179, 8, 0.5)', 
        backgroundColor: theme.colors.warning.subtle, 
        color: theme.colors.warning.DEFAULT,
        boxShadow: `0 0 15px -3px ${theme.colors.warning.glow}`
      };
      default: return { 
        borderColor: theme.colors.border, 
        backgroundColor: theme.colors.surface.elevated, 
        color: theme.colors.text.muted 
      };
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'done': return <CheckCircle2 className="w-3 h-3" />;
      case 'failed': return <AlertCircle className="w-3 h-3" />;
      case 'running': return <Play className="w-3 h-3" />;
      case 'awaiting_approval': return <ShieldCheck className="w-3 h-3" />;
      default: return <Clock className="w-3 h-3" />;
    }
  };

  const styles = getStatusStyles(data.status);

  return (
    <div 
      className={`px-4 py-2 rounded-xl border min-w-[150px] shadow-xl backdrop-blur-sm transition-all`}
      style={styles}
    >
      <Handle type="target" position={Position.Top} className="!bg-white/20 !border-none" />
      <div className="flex flex-col gap-1">
        <div className="flex items-center justify-between gap-2">
          <span className="text-[10px] font-bold uppercase tracking-widest opacity-60">{data.capability}</span>
          {getStatusIcon(data.status)}
        </div>
        <div className="text-xs font-medium line-clamp-1">{data.label}</div>
        {data.status === 'awaiting_approval' && (
          <button 
            onClick={(e) => {
              e.stopPropagation();
              data.onApprove(data.id);
            }}
            className="mt-2 py-1 px-2 text-[10px] font-bold rounded uppercase transition-colors"
            style={{ backgroundColor: theme.colors.warning.DEFAULT, color: '#000' }}
          >
            Approve Action
          </button>
        )}
      </div>
      <Handle type="source" position={Position.Bottom} className="!bg-white/20 !border-none" />
    </div>
  );
};

const nodeTypes = {
  custom: CustomNode,
};

const DAGView = ({ plan, onApprove }: { plan: any[], taskID: string, onApprove: (id: string) => void }) => {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  useEffect(() => {
    if (!plan) return;

    const newNodes = plan.map((step, index) => ({
      id: step.step_id,
      type: 'custom',
      position: { x: 250, y: index * 120 },
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
        style: { stroke: theme.colors.primary.DEFAULT, strokeWidth: 2, opacity: 0.4 },
      }))
    );

    setNodes(newNodes);
    setEdges(newEdges);
  }, [plan, setNodes, setEdges, onApprove]);

  return (
    <div className="h-full w-full rounded-2xl overflow-hidden border" style={{ backgroundColor: theme.colors.background, borderColor: theme.colors.border }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        fitView
      >
        <Background color="#111" gap={20} />
        <Controls className="!bg-white/5 !border-white/10 !fill-white" />
      </ReactFlow>
    </div>
  );
};

export default DAGView;
