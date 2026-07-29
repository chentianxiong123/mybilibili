import React from 'react';

function DisabledElement() {
  return React.createElement('div');
}

DisabledElement.mins = { x: 0, y: 0 };

export const Minifeed = DisabledElement;
export const LegacyEvents = DisabledElement;
export { SceneSelectorElement as Scenes } from './SceneSelector';
export { SourceSelectorElement as Sources } from './SourceSelector';
export { Mixer } from './Mixer';
export { RecordingPreview } from './RecordingPreview';
export { StreamPreview } from './StreamPreview';
export { Browser } from './Browser';
export { Display } from './Display';
