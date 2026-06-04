import React, { useEffect, useState } from 'react';
import { Layout, Tooltip } from 'antd';
import Fuse from 'fuse.js';
import moment from 'moment';
import cx from 'classnames';
import { ModalLayout } from 'components-react/shared/ModalLayout';
import { Services } from 'components-react/service-provider';
import { confirmAsync, promptAsync, alertAsync } from 'components-react/modals';
import Scrollable from 'components-react/shared/Scrollable';
import { $t } from 'services/i18n';
import { $i } from 'services/utils';
import { ISceneCollectionsManifestEntry } from 'services/scene-collections';
import { getOS, OS } from 'util/operating-systems';
import styles from './ManageSceneCollections.m.less';
import { TextInput } from 'components-react/shared/inputs';
import { useVuex } from 'components-react/hooks';
import Translate from 'components-react/shared/Translate';
import ButtonHighlighted from 'components-react/shared/ButtonHighlighted';
import * as remote from '@electron/remote';

const { Sider, Content } = Layout;

export default function ManageSceneCollections() {
  const {
    WindowsService,
    SceneCollectionsService,
    ObsImporterService,
        NavigationService,
    UsageStatisticsService,
    UserService,
  } = Services;
  const [query, setQuery] = useState('');

  const { collections, isLoggedIn } = useVuex(() => ({
    collections: SceneCollectionsService.collections,
    isLoggedIn: UserService.views.isLoggedIn,
    
  }));

  function close() {
    SceneCollectionsService.stateService.flushManifestFile();
    WindowsService.actions.closeChildWindow();
  }

  async function create() {
    const name = await promptAsync(
      { title: $t('Enter a Scene Collection Name'), closable: true },
      SceneCollectionsService.suggestName('Scenes'),
    );
    SceneCollectionsService.actions.create({ name });
  }

  function importFromObs() {
    ObsImporterService.actions.import();
  }

  

function CollectionNode(p: {
  collection: ISceneCollectionsManifestEntry;
  recentlyUpdated: boolean;
}) {
  const { SceneCollectionsService } = Services;
  const [duplicating, setDuplicating] = useState(false);
  const modified = moment(p.collection.modified).fromNow();
  const isActive = p.collection.id === SceneCollectionsService.activeCollection?.id;

  useEffect(onNeedsRenamedChanged, [p.collection.needsRename]);

  function onNeedsRenamedChanged() {
    if (p.collection.needsRename) rename();
  }

  async function makeActive() {
    if (p.collection.operatingSystem !== getOS()) return;
    SceneCollectionsService.actions.load(p.collection.id);
  }

  function duplicate() {
    setDuplicating(true);

    setTimeout(() => {
      SceneCollectionsService.actions.return
        .duplicate(p.collection.name, p.collection.id)
        .finally(() => setDuplicating(false));
    }, 500);
  }

  async function rename() {
    const newName = await promptAsync(
      { title: $t('Enter a Scene Collection Name'), closable: true },
      p.collection.name,
    );
    SceneCollectionsService.actions.rename(newName, p.collection.id);
  }

  async function remove() {
    const deleteConfirmed = await confirmAsync(
      $t('Are you sure you want to remove %{collectionName}?', {
        collectionName: p.collection.name,
      }),
    );
    if (deleteConfirmed) SceneCollectionsService.actions.delete(p.collection.id);
  }

  return (
    <div
      onDoubleClick={makeActive}
      className={cx(styles.collectionNode, { [styles.active]: isActive })}
    >
      <span>
        <i
          className={cx(
            'fab',
            p.collection.operatingSystem === OS.Windows ? 'fa-windows' : 'fa-apple',
          )}
        />
        {p.collection.name}
      </span>
      {p.recentlyUpdated && <span className={styles.whisper}>Updated {modified}</span>}
      <div className={styles.editIcons}>
        <Tooltip title={$t('Rename')}>
          <i className="icon-edit" onClick={rename} />
        </Tooltip>
        {!duplicating && (
          <Tooltip title={$t('Duplicate')}>
            <i className="icon-copy" onClick={duplicate} />
          </Tooltip>
        )}
        {duplicating && <i className="fa fa-spinner fa-pulse" />}
        <Tooltip title={$t('Delete')}>
          <i className="icon-trash" onClick={remove} />
        </Tooltip>
      </div>
    </div>
  );
}
