import React from 'react';
import { EDismissable } from 'services/dismissables';

interface IHelpTipProps {
  title: string;
  dismissableKey: EDismissable;
  position: {
    top?: string;
    left?: string;
    bottom?: string;
    right?: string;
  };
  tipPosition?: 'left' | 'right';
  arrowPosition?: 'top' | 'bottom';
  style?: React.CSSProperties;
}

export default function HelpTip(_props: React.PropsWithChildren<IHelpTipProps>) {
  return null;
}
