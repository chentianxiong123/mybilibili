import React, { useEffect, useMemo, useState } from 'react';
import { Checkbox, Input, InputNumber, Select } from 'antd';
import { Services } from '../service-provider';
import { TObsFormData } from '../../components/obs/inputs/ObsInput';
import { ModalLayout } from '../shared/ModalLayout';
import Display from '../shared/Display';
import { useSubscription } from '../hooks/useSubscription';
import { useChildWindowParams } from 'components-react/hooks';
import { $t } from 'services/i18n';
import { Input as ObsInput } from 'services/obs-api';

export default function SourceProperties() {
  const { WindowsService, SourcesService, EditorCommandsService } = Services;
  const sourceId = useChildWindowParams('sourceId');
  const source = useMemo(() => SourcesService.views.getSource(sourceId), [sourceId]);
  const [properties, setProperties] = useState<TObsFormData>(() => source?.getPropertiesFormData() || []);

  // close the window if the source has been deleted
  useSubscription(SourcesService.sourceRemoved, removedSource => {
    if (source && removedSource.sourceId !== source.sourceId) return;
    WindowsService.actions.closeChildWindow();
  });

  // update properties state if the source has been changed
  useSubscription(SourcesService.sourceUpdated, updatedSource => {
    if (source && updatedSource.sourceId !== source.sourceId) return;
    setProperties(source?.getPropertiesFormData() || []);
  });

  useEffect(() => {
    setProperties(source?.getPropertiesFormData() || []);
  }, [sourceId]);

  function updateProperty(index: number, value: unknown) {
    if (!source) return;
    const next = [...properties];
    next[index] = { ...next[index], value } as TObsFormData[number];
    setProperties(next);
    EditorCommandsService.actions.executeCommand('EditSourcePropertiesCommand', source.sourceId, [
      next[index],
    ]);
  }

  if (!source) return <ModalLayout><div>{$t('Source not found')}</div></ModalLayout>;
  if (!properties.length) {
    return (
      <ModalLayout
        scrollable
        fixedChild={<Display sourceId={source.sourceId} style={{ position: 'relative' }} />}
      >
        <div style={{ padding: 20 }}>{$t('No settings available for this source')}</div>
      </ModalLayout>
    );
  }

  return (
    <ModalLayout
      scrollable
      fixedChild={<Display sourceId={source.sourceId} style={{ position: 'relative' }} />}
    >
      <div style={{ display: 'grid', gap: 12, padding: 20 }}>
        {properties.map((prop, index) => {
          if (prop.visible === false) return null;
          const common = { key: prop.name, style: { width: '100%' } };
          if (prop.type === 'OBS_PROPERTY_BOOL') {
            return (
              <label key={prop.name} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Checkbox checked={!!prop.value} onChange={e => updateProperty(index, e.target.checked)} />
                <span>{prop.description || prop.name}</span>
              </label>
            );
          }
          if (prop.type === 'OBS_PROPERTY_LIST') {
            return (
              <div key={prop.name}>
                <div style={{ marginBottom: 6 }}>{prop.description || prop.name}</div>
                <Select
                  value={prop.value as string}
                  style={{ width: '100%' }}
                  onChange={value => updateProperty(index, value)}
                  options={prop.options.map(opt => ({ label: opt.description, value: opt.value as any }))}
                />
              </div>
            );
          }
          if (prop.type === 'OBS_PROPERTY_INT' || prop.type === 'OBS_PROPERTY_FLOAT' || prop.type === 'OBS_PROPERTY_DOUBLE' || prop.type === 'OBS_PROPERTY_SLIDER') {
            return (
              <div key={prop.name}>
                <div style={{ marginBottom: 6 }}>{prop.description || prop.name}</div>
                <InputNumber
                  {...common}
                  value={Number(prop.value)}
                  min={(prop as any).minVal}
                  max={(prop as any).maxVal}
                  step={(prop as any).stepVal}
                  onChange={value => updateProperty(index, value)}
                />
              </div>
            );
          }
          if (prop.type === 'OBS_PROPERTY_COLOR') {
            return (
              <div key={prop.name}>
                <div style={{ marginBottom: 6 }}>{prop.description || prop.name}</div>
                <Input value={String(prop.value ?? '')} onChange={e => updateProperty(index, e.target.value)} />
              </div>
            );
          }
          if (prop.type === 'OBS_PROPERTY_FILE' || prop.type === 'OBS_PROPERTY_PATH') {
            return (
              <div key={prop.name}>
                <div style={{ marginBottom: 6 }}>{prop.description || prop.name}</div>
                <Input value={String(prop.value ?? '')} onChange={e => updateProperty(index, e.target.value)} />
              </div>
            );
          }
          return (
            <div key={prop.name}>
              <div style={{ marginBottom: 6 }}>{prop.description || prop.name}</div>
              <Input.TextArea
                value={String(prop.value ?? '')}
                onChange={e => updateProperty(index, e.target.value)}
                autoSize={{ minRows: 1, maxRows: 4 }}
              />
            </div>
          );
        })}
      </div>
    </ModalLayout>
  );
}
