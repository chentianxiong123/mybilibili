import React from 'react';
import MybilibiliLive from './pages/MybilibiliLive';
import Loader from './pages/Loader';
import Display from './shared/Display';
import OriginalMain from './windows/Main';
import AddSource from './windows/AddSource';
import AdvancedAudio from './windows/advanced-audio';
import AdvancedStatistics from './windows/AdvancedStatistics';
import EditTransform from './windows/EditTransform';
import NameFolder from './windows/NameFolder';
import NameScene from './windows/NameScene';
import Projector from './windows/Projector';
import SourceFilters from './windows/SourceFilters';
import SourceProperties from './windows/SourceProperties';
import SourceShowcase from './windows/source-showcase';
import { Services } from './service-provider';
import { useVuex } from './hooks';

const el = React.createElement;

const settingsStyles: Record<string, React.CSSProperties> = {
  frame: {
    display: 'grid',
    gridTemplateColumns: '220px minmax(0, 1fr)',
    width: '100%',
    height: '100%',
    background: '#171717',
    color: '#ede9df',
    fontFamily: 'Inter, Arial, sans-serif',
    overflow: 'hidden',
  },
  side: {
    borderRight: '1px solid #34322c',
    background: '#1f1e1b',
    overflowY: 'auto',
    padding: 12,
  },
  title: {
    margin: '4px 6px 14px',
    color: '#fff',
    fontSize: 16,
    fontWeight: 800,
  },
  category: {
    display: 'block',
    width: '100%',
    minHeight: 34,
    marginBottom: 4,
    padding: '0 10px',
    border: '1px solid transparent',
    borderRadius: 6,
    background: 'transparent',
    color: '#d6cec1',
    fontSize: 13,
    fontWeight: 700,
    textAlign: 'left',
    cursor: 'pointer',
  },
  categoryActive: {
    borderColor: '#1fbf75',
    background: '#14261d',
    color: '#fff',
  },
  content: {
    minWidth: 0,
    overflowY: 'auto',
    padding: 18,
  },
  contentHeader: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 12,
    marginBottom: 14,
  },
  h1: {
    margin: 0,
    color: '#fff',
    fontSize: 20,
    lineHeight: '28px',
  },
  button: {
    minHeight: 32,
    border: '1px solid #3b3830',
    borderRadius: 6,
    background: '#24231f',
    color: '#ede9df',
    padding: '0 12px',
    fontSize: 13,
    fontWeight: 700,
    cursor: 'pointer',
  },
  group: {
    marginBottom: 14,
    border: '1px solid #34322c',
    borderRadius: 8,
    background: '#201f1c',
    overflow: 'hidden',
  },
  groupTitle: {
    padding: '10px 12px',
    borderBottom: '1px solid #34322c',
    color: '#fff',
    fontSize: 14,
    fontWeight: 800,
  },
  row: {
    display: 'grid',
    gridTemplateColumns: 'minmax(180px, 0.55fr) minmax(220px, 1fr)',
    gap: 12,
    alignItems: 'center',
    padding: '9px 12px',
    borderBottom: '1px solid #302e28',
  },
  label: {
    minWidth: 0,
    overflow: 'hidden',
    color: '#c8c1b4',
    fontSize: 13,
    fontWeight: 700,
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  input: {
    width: '100%',
    minWidth: 0,
    height: 32,
    border: '1px solid #3b3830',
    borderRadius: 6,
    background: '#181715',
    color: '#ede9df',
    padding: '0 10px',
    fontSize: 13,
  },
  code: {
    display: 'block',
    minWidth: 0,
    overflow: 'hidden',
    color: '#a8a092',
    fontFamily: 'Consolas, monospace',
    fontSize: 12,
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  empty: {
    display: 'flex',
    height: '100%',
    alignItems: 'center',
    justifyContent: 'center',
    color: '#a8a092',
    fontSize: 14,
  },
};

function settingValueText(value: unknown) {
  if (value == null) return '';
  if (typeof value === 'object') {
    try {
      return JSON.stringify(value);
    } catch (e: unknown) {
      return String(value);
    }
  }
  return String(value);
}

function optionValue(value: unknown) {
  return value == null ? '__MYBILIBILI_NULL__' : String(value);
}

function Settings() {
  const { NavigationService, SettingsService } = Services;
  const settingsState = useVuex(() => ({
    currentTab: NavigationService.state.currentSettingsTab,
    settings: SettingsService.state,
  }));
  const categories = Object.keys(settingsState.settings).filter(category => {
    const lower = category.toLowerCase();
    return (
      category !== 'StreamSecond' &&
      !lower.includes('prime') &&
      !lower.includes('store') &&
      !lower.includes('donation') &&
      !lower.includes('subscription')
    );
  });
  const [selected, setSelected] = React.useState(
    settingsState.currentTab && categories.includes(settingsState.currentTab)
      ? settingsState.currentTab
      : categories[0] || '',
  );

  React.useEffect(() => {
    if (selected && categories.includes(selected)) return;
    const nextCategory =
      settingsState.currentTab && categories.includes(settingsState.currentTab)
        ? settingsState.currentTab
        : categories[0] || '';
    setSelected(nextCategory);
  }, [categories.join('|'), settingsState.currentTab]);

  function selectCategory(category: string) {
    setSelected(category);
    NavigationService.actions.setSettingsNavigation(category as any);
  }

  function saveSetting(category: string, name: string, value: unknown) {
    SettingsService.actions.setSettingValue(category as any, name, value as any);
  }

  function renderInput(category: string, setting: any) {
    if (setting.visible === false) return null;
    const disabled = setting.enabled === false;
    const label = setting.description || setting.name;

    if (Array.isArray(setting.options)) {
      return el(
        'div',
        { key: setting.name, style: settingsStyles.row },
        el('span', { style: settingsStyles.label, title: label }, label),
        el(
          'select',
          {
            style: settingsStyles.input,
            value: optionValue(setting.value),
            disabled,
            onChange: (event: React.ChangeEvent<HTMLSelectElement>) => {
              const option = setting.options.find(
                (item: any) => optionValue(item.value) === event.target.value,
              );
              saveSetting(category, setting.name, option ? option.value : event.target.value);
            },
          },
          setting.options.map((option: any) =>
            el(
              'option',
              { key: optionValue(option.value), value: optionValue(option.value) },
              option.description || settingValueText(option.value),
            ),
          ),
        ),
      );
    }

    if (typeof setting.value === 'boolean' || setting.type === 'OBS_PROPERTY_BOOL') {
      return el(
        'div',
        { key: setting.name, style: settingsStyles.row },
        el('span', { style: settingsStyles.label, title: label }, label),
        el('input', {
          type: 'checkbox',
          checked: !!setting.value,
          disabled,
          onChange: (event: React.ChangeEvent<HTMLInputElement>) =>
            saveSetting(category, setting.name, event.target.checked),
        }),
      );
    }

    if (typeof setting.value === 'number') {
      return el(
        'div',
        { key: setting.name, style: settingsStyles.row },
        el('span', { style: settingsStyles.label, title: label }, label),
        el('input', {
          key: settingValueText(setting.value),
          type: 'number',
          style: settingsStyles.input,
          defaultValue: setting.value,
          disabled,
          onBlur: (event: React.FocusEvent<HTMLInputElement>) =>
            saveSetting(category, setting.name, Number(event.target.value)),
        }),
      );
    }

    if (typeof setting.value === 'object' && setting.value != null) {
      return el(
        'div',
        { key: setting.name, style: settingsStyles.row },
        el('span', { style: settingsStyles.label, title: label }, label),
        el('code', { style: settingsStyles.code, title: settingValueText(setting.value) }, settingValueText(setting.value)),
      );
    }

    return el(
      'div',
      { key: setting.name, style: settingsStyles.row },
      el('span', { style: settingsStyles.label, title: label }, label),
      el('input', {
        key: settingValueText(setting.value),
        style: settingsStyles.input,
        defaultValue: settingValueText(setting.value),
        disabled,
        onBlur: (event: React.FocusEvent<HTMLInputElement>) =>
          saveSetting(category, setting.name, event.target.value),
      }),
    );
  }

  const category = settingsState.settings[selected];

  return el(
    'div',
    { style: settingsStyles.frame },
    el(
      'aside',
      { style: settingsStyles.side },
      el('h1', { style: settingsStyles.title }, '设置'),
      categories.map(categoryName =>
        el(
          'button',
          {
            key: categoryName,
            style:
              categoryName === selected
                ? { ...settingsStyles.category, ...settingsStyles.categoryActive }
                : settingsStyles.category,
            onClick: () => selectCategory(categoryName),
          },
          categoryName,
        ),
      ),
    ),
    el(
      'main',
      { style: settingsStyles.content },
      category
        ? [
            el(
              'div',
              { key: 'header', style: settingsStyles.contentHeader },
              el('h1', { style: settingsStyles.h1 }, selected),
              el(
                'button',
                {
                  style: settingsStyles.button,
                  onClick: () => SettingsService.actions.loadSettingsIntoStore(),
                },
                '刷新',
              ),
            ),
            ...category.formData.map((group: any, index: number) =>
              el(
                'section',
                { key: `${group.nameSubCategory}-${index}`, style: settingsStyles.group },
                el(
                  'div',
                  { style: settingsStyles.groupTitle },
                  group.nameSubCategory || '设置',
                ),
                group.parameters.map((setting: any) => renderInput(selected, setting)),
              ),
            ),
          ]
        : el('div', { style: settingsStyles.empty }, '设置尚未加载'),
    ),
  );
}

function Main() {
  return el(OriginalMain);
}

function Blank() {
  return null;
}

function TitleBar(props: { windowId?: string }) {
  return el(
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
  return el('div');
}

export const components = {
  Main,
  Loader,
  Blank,
  TitleBar,
  Display,
  BrowserView: Placeholder,
  Settings,
  TestWidgets: Placeholder,
  AddSource,
  AdvancedAudio,
  AdvancedStatistics,
  EditTransform,
  NameFolder,
  NameScene,
  Projector,
  SourceFilters,
  SourceProperties,
  SourceShowcase,
  MybilibiliLive,
};
