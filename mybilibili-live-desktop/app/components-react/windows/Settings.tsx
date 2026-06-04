import React, { useEffect, useMemo, useState } from 'react';
import { MenuInfo } from 'rc-menu/lib/interface';
import * as pages from './settings/pages';
import { ESettingsCategory, TCategoryName } from 'services/settings';
import { EDismissable } from 'services/dismissables';
import { Services } from 'components-react/service-provider';
import { useRealmObject } from 'components-react/hooks/realm';
import { useVuex } from 'components-react/hooks';
import { $t } from 'services/i18n';
import { ModalLayout } from 'components-react/shared/ModalLayout';
import { Menu } from 'antd';
import Scrollable from 'components-react/shared/Scrollable';
import styles from './Settings.m.less';
import { TextInput } from 'components-react/shared/inputs';
import DismissableBadge from 'components-react/shared/DismissableBadge';
import SearchablePages from 'components-react/shared/SearchablePages';

export interface ISettingsProps {
  globalSearchStr: string;
}

interface ISettingsConfig {
  icon: string;
  component: React.FunctionComponent<ISettingsProps>;
  dismissable?: EDismissable;
  shouldShow?: () => boolean;
}

export const SETTINGS_CONFIG: Record<ESettingsCategory, ISettingsConfig> = {
  [ESettingsCategory.General]: { icon: 'icon-overview', component: pages.GeneralSettings },
  [ESettingsCategory.Stream]: { icon: 'fas fa-globe', component: pages.StreamSettings },
  [ESettingsCategory.Output]: { icon: 'fas fa-microchip', component: pages.OutputSettings },
  [ESettingsCategory.Video]: { icon: 'fas fa-film', component: pages.VideoSettings },
  [ESettingsCategory.Audio]: { icon: 'icon-audio', component: pages.AudioSettings },
  [ESettingsCategory.Hotkeys]: { icon: 'icon-settings', component: pages.Hotkeys },
  [ESettingsCategory.GameOverlay]: {
    icon: 'icon-full-screen',
    component: pages.GameOverlay,
    shouldShow: () => false,
  },
  [ESettingsCategory.VirtualWebcam]: {
    icon: 'fas fa-camera',
    component: pages.VirtualWebcamSettings,
  },
  [ESettingsCategory.Advanced]: { icon: 'fas fa-cogs', component: pages.AdvancedSettings },
  [ESettingsCategory.Developer]: {
    icon: 'far fa-file-code',
    component: pages.DeveloperSettings,
    shouldShow: () => false,
  },
  [ESettingsCategory.SceneCollections]: {
    icon: 'icon-主题',
    component: pages.SceneCollectionsSettings,
  },
  [ESettingsCategory.Notifications]: {
    icon: 'icon-notifications',
    component: pages.NotificationSettings,
  },
  [ESettingsCategory.Appearance]: {
    icon: 'icon-settings-3-1',
    component: pages.AppearanceSettings,
    dismissable: EDismissable.CustomMenuSettings,
  },
  [ESettingsCategory.Mobile]: {
    icon: 'icon-phone-case',
    component: pages.MobileSettings,
    shouldShow: () => false,
  },
  [ESettingsCategory.Experimental]: {
    icon: 'fas fa-flask',
    component: pages.ExperimentalSettings,
    shouldShow: () => false,
  },
  [ESettingsCategory.InstalledApps]: {
    icon: 'icon-store',
    component: pages.InstalledApps,
    shouldShow: () => false,
  },
  [ESettingsCategory.GetSupport]: {
    icon: 'icon-question',
    component: pages.Support,
    shouldShow: () => false,
  },
  [ESettingsCategory.AI]: {
    icon: 'icon-ai',
    component: pages.AISettings,
    shouldShow: () => false,
  },
};

const SETTINGS_CATEGORY_LABELS: Partial<Record<ESettingsCategory, string>> = {
  [ESettingsCategory.General]: '通用',
  [ESettingsCategory.Stream]: '直播',
  [ESettingsCategory.Output]: '输出',
  [ESettingsCategory.Audio]: '音频',
  [ESettingsCategory.Video]: '视频',
  [ESettingsCategory.Hotkeys]: '快捷键',
  [ESettingsCategory.Advanced]: '高级',
  [ESettingsCategory.SceneCollections]: '场景集合',
  [ESettingsCategory.Notifications]: '通知',
  [ESettingsCategory.Appearance]: '外观',
  [ESettingsCategory.VirtualWebcam]: '虚拟摄像头',
  [ESettingsCategory.Developer]: '开发者',
  [ESettingsCategory.Experimental]: '实验功能',
};

function settingsCategoryLabel(category: ESettingsCategory) {
  return SETTINGS_CATEGORY_LABELS[category] || $t(category);
}

export default function Settings() {
  const {
    SettingsService,
    NavigationService,
    DismissablesService,
  } = Services;

  const currentTab = useRealmObject(NavigationService.state).currentSettingsTab;

  const { showDismissable } = useVuex(() => ({
    showDismissable: (value: EDismissable) => DismissablesService.views.shouldShow(value),
  }));

  const categories = useMemo(
    () =>
      SettingsService.getCategories().filter(category => {
        const config = SETTINGS_CONFIG[category];
        return !config.shouldShow || config.shouldShow();
      }),
    [],
  );
  const activeTab = categories.includes(currentTab as ESettingsCategory)
    ? currentTab
    : ESettingsCategory.General;

  useEffect(() => {
    // Make sure we have the latest settings
    SettingsService.actions.loadSettingsIntoStore();
  }, []);

  useEffect(() => {
    if (activeTab !== currentTab) {
      NavigationService.actions.setSettingsNavigation(activeTab);
    }
  }, [activeTab, currentTab]);

  function setCurrentTab(value: TCategoryName) {
    if (!value) return;
    NavigationService.actions.setSettingsNavigation(value);
  }

  function handleMenuNavigation(event: MenuInfo) {
    setCurrentTab(event.key as TCategoryName);
  }

  function dismiss(category: TCategoryName) {
    const dismissable = SETTINGS_CONFIG[category].dismissable;
    if (dismissable) DismissablesService.actions.dismiss(dismissable);
  }

  /** PAGE SEARCH LOGIC */
  const [searchStr, setSearchStr] = useState('');

  function handleSearchCompleted(foundPages: TCategoryName[]) {
    const filteredPages = foundPages.filter(page => categories.includes(page as ESettingsCategory));
    setCurrentTab(filteredPages[0]);
  }

  function onSearchInput(str: string) {
    setSearchStr(str);
  }

  const SettingsContent = SETTINGS_CONFIG[activeTab].component;

  return (
    <ModalLayout bodyClassName={styles.settings}>
      <Scrollable className={styles.settingsNav}>
        <div style={{ padding: '0 24px 12px 24px' }}>
          <TextInput
            prefix={<i className="icon-search" />}
            placeholder="搜索设置"
            value={searchStr}
            onChange={onSearchInput}
            uncontrolled={false}
            nowrap
          />
        </div>
        <Menu
          mode="inline"
          selectedKeys={[activeTab]}
          onClick={handleMenuNavigation}
          style={{ paddingBottom: '46px' }}
        >
          {categories.map(cat => {
            const config = SETTINGS_CONFIG[cat];
            if (!config.shouldShow || config.shouldShow()) {
              return (
                <Menu.Item key={cat} icon={<i className={config.icon} />}>
                  <div
                    style={{ display: 'flex' }}
                    onClick={() => dismiss(cat)}
                    data-name="settings-nav-item"
                  >
                    {settingsCategoryLabel(cat)}
                    {config.dismissable && showDismissable(config.dismissable) && (
                      <DismissableBadge dismissableKey={config.dismissable} />
                    )}
                  </div>
                </Menu.Item>
              );
            }
          })}
        </Menu>
      </Scrollable>
      <Scrollable className={styles.settingsContainer} snapToWindowEdge>
        <div className={styles.settingsContent}>
          <SearchablePages
            onSearchCompleted={handleSearchCompleted}
            pages={categories}
            page={activeTab}
            searchStr={searchStr}
          >
            <SettingsContent globalSearchStr={searchStr} />
          </SearchablePages>
        </div>
      </Scrollable>
    </ModalLayout>
  );
}
