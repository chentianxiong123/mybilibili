import React, { useEffect, useMemo, useState } from 'react';
import { Button, Divider, List, Modal, Select } from 'antd';
import { Services } from 'components-react/service-provider';
import { ModalLayout } from 'components-react/shared/ModalLayout';
import Display from 'components-react/shared/Display';
import Scrollable from 'components-react/shared/Scrollable';
import { useChildWindowParams, useVuex } from 'components-react/hooks';
import { TextInput } from 'components-react/shared/inputs';
import Form, { useForm } from 'components-react/shared/inputs/Form';
import { TSourceFilterType } from 'services/source-filters';
import { $t } from 'services/i18n';

export default function SourceFilters() {
  const { WindowsService, SourceFiltersService, SourcesService, EditorCommandsService } = Services;
  const sourceId = useChildWindowParams('sourceId');
  const { filters, source, presetValue, presetOptions, types } = useVuex(() => ({
    filters: SourceFiltersService.getFilters(sourceId),
    source: SourcesService.views.getSource(sourceId),
    presetValue: SourceFiltersService.views.presetFilterBySourceId(sourceId)?.settings
      ?.image_path || 'none',
    presetOptions: SourceFiltersService.views.presetFilterOptionsReact,
    types: SourceFiltersService.views.getTypesForSource(sourceId),
  }));

  const [selectedFilter, setSelectedFilter] = useState(filters?.[0]?.name || '');
  const [createModal, setCreateModal] = useState(false);

  useEffect(() => {
    if (!filters || filters.length === 0) {
      setSelectedFilter('');
      return;
    }
    if (!selectedFilter || !filters.find(filter => filter.name === selectedFilter)) {
      setSelectedFilter(filters[0].name);
    }
  }, [filters, selectedFilter]);

  if (!source) {
    return <ModalLayout><div>{$t('Source not found')}</div></ModalLayout>;
  }

  return (
    <ModalLayout fixedChild={<Display sourceId={sourceId} />} bodyStyle={{ padding: 0 }}>
      <div style={{ display: 'flex', height: '100%' }}>
        <div style={{ width: 280, borderRight: '1px solid var(--border)', padding: 16 }}>
          <Select
            value={presetValue}
            onChange={value => {
              if (value === 'none') {
                SourceFiltersService.actions.removePresetFilter(sourceId);
              } else {
                SourceFiltersService.actions.addPresetFilter(sourceId, value);
              }
            }}
            options={presetOptions}
            style={{ width: '100%', marginBottom: 12 }}
          />
          <Button block onClick={() => setCreateModal(true)}>
            {$t('Add Filter')}
          </Button>
          <Divider />
          <Scrollable style={{ height: 'calc(100% - 100px)' }}>
            <List
              dataSource={filters || []}
              locale={{ emptyText: $t('No filters') }}
              renderItem={filter => (
                <List.Item
                  actions={[
                    <Button
                      key="toggle"
                      type="link"
                      onClick={() =>
                        SourceFiltersService.actions.setVisibility(
                          sourceId,
                          filter.name,
                          !filter.visible,
                        )
                      }
                    >
                      {filter.visible ? $t('Hide') : $t('Show')}
                    </Button>,
                    <Button
                      key="remove"
                      type="link"
                      danger
                      onClick={() => SourceFiltersService.actions.remove(sourceId, filter.name)}
                    >
                      {$t('Remove')}
                    </Button>,
                  ]}
                  onClick={() => setSelectedFilter(filter.name)}
                  style={{ cursor: 'pointer' }}
                >
                  {filter.name}
                </List.Item>
              )}
            />
          </Scrollable>
        </div>
        <div style={{ flex: 1, padding: 16 }}>
          {selectedFilter ? (
            <FilterEditor
              sourceId={sourceId}
              filterName={selectedFilter}
              onSave={formData => SourceFiltersService.actions.setPropertiesFormData(sourceId, selectedFilter, formData)}
            />
          ) : (
            <div>{$t('Select a filter to edit')}</div>
          )}
        </div>
      </div>
      <CreateFilterModal
        visible={createModal}
        sourceId={sourceId}
        types={types}
        onClose={() => setCreateModal(false)}
        onSubmit={name => {
          setCreateModal(false);
          setSelectedFilter(name);
        }}
      />
    </ModalLayout>
  );
}

function FilterEditor(p: {
  sourceId: string;
  filterName: string;
  onSave: (formData: any[]) => void;
}) {
  const { SourceFiltersService } = Services;
  const formData = useMemo(
    () => SourceFiltersService.getPropertiesFormData(p.sourceId, p.filterName),
    [p.sourceId, p.filterName],
  );
  const [state, setState] = useState(formData);

  useEffect(() => {
    setState(formData);
  }, [formData]);

  if (!state || state.length === 0) {
    return <div>{$t('No settings are available for this filter')}</div>;
  }

  return (
    <div style={{ display: 'grid', gap: 12 }}>
      {state.map((item, index) => (
        <div key={item.name}>
          <div style={{ marginBottom: 6 }}>{item.description || item.name}</div>
          <TextInput
            value={String(item.value ?? '')}
            onInput={value => {
              const next = [...state];
              next[index] = { ...next[index], value };
              setState(next);
              p.onSave(next as any[]);
            }}
            uncontrolled={false}
            layout="vertical"
          />
        </div>
      ))}
    </div>
  );
}

function CreateFilterModal(p: {
  visible: boolean;
  sourceId: string;
  types: { type: TSourceFilterType; description: string }[];
  onClose: () => void;
  onSubmit: (filterName: string) => void;
}) {
  const { SourceFiltersService, EditorCommandsService } = Services;
  const form = useForm();
  const [type, setType] = useState<TSourceFilterType>(p.types[0]?.type);
  const [name, setName] = useState(
    SourceFiltersService.views.suggestName(p.sourceId, p.types[0]?.description || 'Filter'),
  );

  useEffect(() => {
    if (!p.types.length) return;
    setType(p.types[0].type);
    setName(SourceFiltersService.views.suggestName(p.sourceId, p.types[0].description || 'Filter'));
  }, [p.sourceId, p.types]);

  if (!p.types.length) {
    return (
      <Modal visible={p.visible} onCancel={p.onClose} footer={null} getContainer={false}>
        <div>{$t('No filters are available for this source')}</div>
      </Modal>
    );
  }

  async function submit() {
    try {
      await form.validateFields();
    } catch (e: unknown) {
      return;
    }
    //@ts-ignore
    EditorCommandsService.actions.return.executeCommand('AddFilterCommand', p.sourceId, type, name).then(() => p.onSubmit(name));
  }

  return (
    <Modal visible={p.visible} onCancel={p.onClose} footer={null} getContainer={false}>
      <Form form={form} name="createFilterForm" onFinish={submit}>
        <Select
          value={type}
          onChange={value => setType(value)}
          options={p.types.map(t => ({ value: t.type, label: t.description }))}
          style={{ width: '100%', marginBottom: 12 }}
        />
        <TextInput
          value={name}
          onInput={setName}
          label={$t('Filter name')}
          uncontrolled={false}
          layout="vertical"
          required
        />
        <Button type="primary" htmlType="submit">
          {$t('Add')}
        </Button>
      </Form>
    </Modal>
  );
}
