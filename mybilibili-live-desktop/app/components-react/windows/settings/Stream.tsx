import React from 'react';
import { EStreamingState } from '../../../services/streaming';
import { $t } from '../../../services/i18n';
import { Services } from '../../service-provider';
import { ObsGenericSettingsForm } from './ObsSettings';

export function StreamSettings() {
  const canEditSettings =
    Services.StreamingService.state.streamingStatus === EStreamingState.Offline;

  return (
    <div className="section">
      {!canEditSettings && (
        <div className="section section--warning">
          {$t("You can not change these settings when you're live")}
        </div>
      )}
      {canEditSettings && <ObsGenericSettingsForm page="Stream" />}
    </div>
  );
}
