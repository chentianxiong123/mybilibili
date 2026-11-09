import React, { memo, useMemo } from 'react';
import cx from 'classnames';
import electron from 'electron';
import Utils from 'services/utils';
import { Menu } from 'antd';
import { EMenuItemKey, ENavName, IMenuItem, menuTitles } from 'services/side-nav';
import { Services } from '../service-provider';
import { useVuex } from '../hooks';
import MenuItem from 'components-react/shared/MenuItem';
import styles from './NavTools.m.less';

export default memo(function NavTools() {
  const { SettingsService, SideNavService } = Services;
  const isDevMode = useMemo(() => Utils.isDevMode(), []);

  const { menuItems, isOpen } = useVuex(
    () => ({
      menuItems: SideNavService.views.state[ENavName.BottomNav].menuItems,
      isOpen: SideNavService.views.isOpen,
    }),
    false,
  );

  function openSettingsWindow() {
    SettingsService.actions.showSettings();
  }

  function openDevTools() {
    electron.ipcRenderer.send('openDevTools');
  }

  return (
    <Menu
      key={ENavName.BottomNav}
      mode="inline"
      className={cx(styles.bottomNav, !isOpen && styles.closed, isOpen && styles.open)}
      getPopupContainer={triggerNode => triggerNode}
    >
      {menuItems.map((menuItem: IMenuItem) => {
        if (menuItem.key === EMenuItemKey.Settings) {
          return <NavToolsItem key={menuItem.key} menuItem={menuItem} onClick={openSettingsWindow} />;
        }
        if (isDevMode && menuItem.key === EMenuItemKey.DevTools) {
          return <NavToolsItem key={menuItem.key} menuItem={menuItem} onClick={openDevTools} />;
        }
        return null;
      })}
    </Menu>
  );
});

function NavToolsItem(p: { menuItem: IMenuItem; onClick: () => void }) {
  const { menuItem, onClick } = p;
  const title = menuTitles(menuItem.key);

  return (
    <MenuItem title={title} icon={<i className={menuItem.icon} />} onClick={onClick}>
      {title}
    </MenuItem>
  );
}
