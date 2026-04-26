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

import React, { useEffect } from 'react';
import { theme } from '@/theme';
import { X } from 'lucide-react';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  subtitle?: string;
  children: React.ReactNode;
  maxWidth?: string;
}

/**
 * Modal is a generic, reusable container for high-level UI interactions.
 * It handles backdrop blur, transitions, and accessibility.
 */
const Modal = ({ isOpen, onClose, title, subtitle, children, maxWidth = 'max-w-xl' }: ModalProps) => {
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => { document.body.style.overflow = 'unset'; };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-6">
      <div 
        className="absolute inset-0 bg-slate-900/40 backdrop-blur-sm animate-in fade-in duration-300" 
        onClick={onClose} 
      />
      
      <div 
        className={`relative w-full ${maxWidth} bg-white shadow-2xl animate-in zoom-in-95 slide-in-from-bottom-4 duration-300 overflow-hidden flex flex-col`}
        style={{ borderRadius: theme.spacing.borderRadius['2xl'] }}
      >
        <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-blue-500 via-indigo-500 to-purple-500" />
        
        <div className="p-8 flex-1 overflow-y-auto premium-scrollbar">
          {(title || subtitle) && (
            <div className="flex items-center justify-between mb-8">
              <div>
                {title && <h2 className="text-xl font-black tracking-tight text-slate-900">{title}</h2>}
                {subtitle && <p className="text-[10px] font-bold text-slate-400 uppercase tracking-widest mt-1">{subtitle}</p>}
              </div>
              <button 
                onClick={onClose}
                className="p-2 hover:bg-slate-50 text-slate-400 hover:text-slate-600 transition-colors"
                style={{ borderRadius: theme.spacing.borderRadius.full }}
              >
                <X className="w-5 h-5" />
              </button>
            </div>
          )}
          
          {children}
        </div>
      </div>
    </div>
  );
};

export default Modal;
