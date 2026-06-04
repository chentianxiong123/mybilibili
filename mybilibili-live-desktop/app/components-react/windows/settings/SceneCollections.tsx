import React, { useState } from 'react';
import { Button } from 'antd';
import { Services } from 'components-react/service-provider';
import { useRealmObject } from 'components-react/hooks/realm';
import { useVuex } from 'components-react/hooks';
import { ObsSettingsSection } from './ObsSettings';
import { CheckboxInput, ListInput } from 'components-react/shared/inputs';

export function SceneCollectionsSettings() {
  const { SceneCollectionsService, CustomizationService } = Services;
  const [collection, setCollection] = useState(SceneCollectionsService.activeCollection?.id || '');
  const { mediaBackupOptOut } = useRealmObject(CustomizationService.state);

  const { collectionOptions } = useVuex(() => ({
    collectionOptions: SceneCollectionsService.collections.map(coll => ({
      label: coll.name,
      value: coll.id,
    })),
  }));

  function setMediaBackupOptOut(value: boolean) {
    CustomizationService.actions.setMediaBackupOptOut(value);
  }

  async function createSceneCollection() {
    const name = await SceneCollectionsService.actions.return.suggestName('直播场景');
    SceneCollectionsService.actions.create({ name });
  }

  return (
    <>
      <ObsSettingsSection>
        <Button onClick={createSceneCollection} type="primary">
          新建场景集合
        </Button>
      </ObsSettingsSection>

      <ObsSettingsSection title="当前场景集合">
        <ListInput
          label="场景集合"
          value={collection}
          onChange={setCollection}
          options={collectionOptions}
        />
      </ObsSettingsSection>

      <ObsSettingsSection>
        <CheckboxInput
          label="不备份本地媒体文件到云端（重启后生效）"
          value={mediaBackupOptOut}
          onChange={setMediaBackupOptOut}
        />
      </ObsSettingsSection>
    </>
  );
}
