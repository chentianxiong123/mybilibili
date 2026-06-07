import React from 'react';
import MybilibiliLive from './pages/MybilibiliLive';
import Loader from './pages/Loader';
import Display from './shared/Display';

function Main() {
  return React.createElement(MybilibiliLive);
}

function Blank() {
  return null;
}

function TitleBar(props: { windowId?: string }) {
  return React.createElement(
    'div',
    {
      style: {
        height: 32,
        display: 'flex',
        alignItems: 'center',
        padding: '0 12px',
        borderBottom: '1px solid #e5e7eb',
        color: '#18191c',
        background: '#ffffff',
        WebkitAppRegion: 'drag',
      },
    },
    `mybilibili 直播工作台${props.windowId ? ` - ${props.windowId}` : ''}`,
  );
}

function Placeholder() {
  return React.createElement('div');
}

export const components = {
  Main,
  Loader,
  Blank,
  TitleBar,
  Display,
  BrowserView: Placeholder,
  Settings: Placeholder,
  TestWidgets: Placeholder,
  Projector: Placeholder,
  SourceProperties: Placeholder,
};
