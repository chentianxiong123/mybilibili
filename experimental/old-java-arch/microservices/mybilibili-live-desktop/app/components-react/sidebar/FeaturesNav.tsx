import React, { memo, useCallback } from 'react';
import cx from 'classnames';
import { Menu, message } from 'antd';
import { ENavName, EMenuItemKey, IMenuItem, menuTitles } from 'services/side-nav';
import { $t } from 'services/i18n';
import { TAppPage } from 'services/navigation';
import { useVuex } from 'components-react/hooks';
import { Services } from 'components-react/service-provider';
import MenuItem from 'components-react/shared/MenuItem';
import styles from './SideNav.m.less';

export default memo(function FeaturesNav() {
  const { NavigationService, SideNavService, TransitionsService, DualOutputService } = Services;

  const { menu, currentMenuItem, setCurrentMenuItem, isOpen, studioMode, dualOutputMode } = useVuex(
    () => ({
      menu: SideNavService.state[ENavName.TopNav],
      currentMenuItem: SideNavService.views.currentMenuItem,
      setCurrentMenuItem: SideNavService.actions.setCurrentMenuItem,
      isOpen: SideNavService.views.isOpen,
      studioMode: TransitionsService.views.studioMode,
      dualOutputMode: DualOutputService.views.dualOutputMode,
    }),
  );

  const navigate = useCallback((page: TAppPage) => {
    NavigationService.actions.navigate(page);
  }, []);

  const handleNavigation = useCallback((menuItem: IMenuItem) => {
    if (menuItem.key === EMenuItemKey.StudioMode) {
      if (dualOutputMode) {
        message.error({
          content: $t('Cannot toggle Studio Mode in Dual Output Mode.'),
          className: styles.toggleError,
        });
        return;
      }
      if (TransitionsService.views.studioMode) {
        TransitionsService.actions.disableStudioMode();
      } else {
        TransitionsService.actions.enableStudioMode();
      }
      return;
    }

    if (menuItem.target) navigate(menuItem.target as TAppPage);
    setCurrentMenuItem(menuItem.key);
  }, [dualOutputMode]);

  return (
    <Menu
      key={ENavName.TopNav}
      mode="inline"
      className={cx(
        styles.topNav,
        isOpen && styles.open,
        !isOpen && styles.siderClosed && styles.closed,
      )}
      defaultSelectedKeys={[currentMenuItem || EMenuItemKey.Editor]}
      getPopupContainer={triggerNode => triggerNode}
    >
      {menu.menuItems
        .filter(menuItem => menuItem.isActive)
        .map(menuItem => {
          const active =
            currentMenuItem === menuItem.key ||
            (menuItem.key === EMenuItemKey.StudioMode && studioMode);

          return (
            <MenuItem
              key={menuItem.key}
              className={cx(!isOpen && styles.closed, active && styles.active)}
              onClick={() => handleNavigation(menuItem as IMenuItem)}
              title={menuTitles(menuItem.key)}
              icon={menuItem.icon && <i className={menuItem.icon} />}
            />
          );
        })}
    </Menu>
  );
});
