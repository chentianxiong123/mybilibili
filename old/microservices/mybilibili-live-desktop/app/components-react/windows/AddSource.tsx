import React, { useMemo, useState } from 'react';
import { Menu } from 'antd';
import cx from 'classnames';
import { Services } from 'components-react/service-provider';
import { useChildWindowParams, useVuex } from 'components-react/hooks';
import { ModalLayout } from 'components-react/shared/ModalLayout';
import Scrollable from 'components-react/shared/Scrollable';
import { ListInput, SwitchInput, TextInput } from 'components-react/shared/inputs';
import Form, { useForm } from 'components-react/shared/inputs/Form';
import Display from 'components-react/shared/Display';
import { $t } from 'services/i18n';
import { ISourceAddOptions, TSourceType } from 'services/sources';
import styles from './AddSource.m.less';

type TSourceTarget = 'new' | 'existing';

export default function AddSource() {
  const { SourcesService, ScenesService, WindowsService, EditorCommandsService } = Services;
  const sourceType = useChildWindowParams('sourceType') as TSourceType;
  const sourceAddOptions = (useChildWindowParams('sourceAddOptions') || {}) as ISourceAddOptions;
  const form = useForm();

  const { activeScene, availableSources, existingSources } = useVuex(() => {
    const activeScene = ScenesService.views.activeScene;
    const availableSources = SourcesService.getAvailableSourcesTypesList().filter(
      source => source.value !== 'scene',
    );
    const existingSources = activeScene
      ? SourcesService.views.getSources().filter(source => {
          if (source.sourceId === activeScene.id) return false;
          return source.type === sourceType || source.type === sourceAddOptions?.sourceId;
        })
      : [];

    return {
      activeScene,
      availableSources,
      existingSources,
    };
  });

  const [mode, setMode] = useState<TSourceTarget>(existingSources.length > 0 ? 'existing' : 'new');
  const [selectedSourceId, setSelectedSourceId] = useState(existingSources[0]?.sourceId || '');
  const [name, setName] = useState(() => SourcesService.views.suggestName(sourceType || $t('Source')));
  const canCreateNew = !['scene'].includes(sourceType) && !!activeScene;

  const sourceList = useMemo(
    () => availableSources.map(source => ({ label: source.description, value: source.value })),
    [availableSources],
  );

  function close() {
    WindowsService.actions.closeChildWindow();
  }

  function addExisting() {
    if (!activeScene || !selectedSourceId) return;
    EditorCommandsService.actions.executeCommand(
      'CreateExistingItemCommand',
      activeScene.id,
      selectedSourceId,
    );
    close();
  }

  async function addNew() {
    if (!activeScene) return;
    try {
      await form.validateFields();
    } catch (e: unknown) {
      return;
    }

    EditorCommandsService.actions.executeCommand(
      'CreateNewItemCommand',
      activeScene.id,
      name,
      sourceType,
      {},
      {
        sourceAddOptions,
        display: 'horizontal',
      },
    );
    close();
  }

  return (
    <ModalLayout
      footer={
        <>
          {canCreateNew && existingSources.length > 0 && (
            <SwitchInput
              label={$t('Add a new source instead')}
              value={mode === 'new'}
              onChange={checked => setMode(checked ? 'new' : 'existing')}
            />
          )}
          <button className="button button--default" onClick={close} style={{ marginRight: 6 }}>
            {$t('Cancel')}
          </button>
          <button
            className="button button--action"
            onClick={mode === 'new' ? addNew : addExisting}
          >
            {$t('Add Source')}
          </button>
        </>
      }
    >
      <div className={styles.container}>
        {!sourceType && <div>{$t('No source selected')}</div>}
        {mode === 'existing' && existingSources.length > 0 && (
          <>
            <div>
              <h4>{$t('Add Existing Source')}</h4>
              <Scrollable className={styles.menuContainer}>
                <Menu
                  mode="vertical"
                  selectedKeys={[selectedSourceId]}
                  onClick={({ key }: { key: string }) => setSelectedSourceId(key)}
                  className={styles.menu}
                >
                  {existingSources.map(source => (
                    <Menu.Item key={source.sourceId}>{source.name}</Menu.Item>
                  ))}
                </Menu>
              </Scrollable>
            </div>
            {selectedSourceId && (
              <Display sourceId={selectedSourceId} style={{ width: 200, height: 200 }} />
            )}
          </>
        )}

        {mode === 'existing' && existingSources.length === 0 && (
          <div>{$t('No existing sources available')}</div>
        )}

        {mode === 'new' && (
          <Form form={form} name="addNewSourceForm" onFinish={addNew}>
            <h4>{$t('Add New Source')}</h4>
            <TextInput
              label={$t('Source name')}
              value={name}
              onInput={setName}
              name="newSourceName"
              autoFocus
              required
              uncontrolled={false}
              layout="vertical"
            />
            <ListInput
              label={$t('Source type')}
              value={sourceType}
              options={sourceList}
              onChange={() => void 0}
              name="sourceType"
              disabled
              layout="vertical"
            />
          </Form>
        )}
      </div>
    </ModalLayout>
  );
}
