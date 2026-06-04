import React from 'react';
import MybilibiliLive from './pages/MybilibiliLive';
import Loader from './pages/Loader';
import Display from './shared/Display';

function Main() {
  return <MybilibiliLive />;
}

function Blank() {
  return null;
}

function TitleBar(props: { windowId?: string }) {
  return (
    <div
      style={{
        height: 32,
        display: 'flex',
        alignItems: 'center',
        padding: '0 12px',
        borderBottom: '1px solid #e5e7eb',
        color: '#18191c',
        background: '#ffffff',
        WebkitAppRegion: 'drag',
      }}
    >
      mybilibili Live Desktop{props.windowId ? ` - ${props.windowId}` : ''}
    </div>
  );
}

function Placeholder() {
  return <div />;
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
  RecentEvents: Placeholder,
  Projector: Placeholder,
  SourceProperties: Placeholder,
};
