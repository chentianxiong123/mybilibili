import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import fs from 'fs';
import * as remote from '@electron/remote';
import cx from 'classnames';
import Animation from 'rc-animate';
import { $t } from 'services/i18n';
import { useDebounce, useVuex } from 'components-react/hooks';
import * as appPages from 'components-react/pages';
import TitleBar from 'components-react/shared/TitleBar';
import { Services } from 'components-react/service-provider';
import SideNav from 'components-react/sidebar/SideNav';
import StudioFooter from 'components-react/root/StudioFooter';
import Loader from 'components-react/pages/Loader';
import antdThemes from 'styles/antd/index';
import { TApplicationTheme } from 'services/customization';
import styles from './Main.m.less';
import { StatefulService } from 'services';
import { useRealmObject } from 'components-react/hooks/realm';

// TODO: this is technically deprecated as we have moved customizationService to Realm
// but some users may still have this value
const loadedTheme = (): TApplicationTheme | undefined => {
  const customizationState = localStorage.getItem('PersistentStatefulService-CustomizationService');
  if (customizationState) {
    return JSON.parse(customizationState)?.theme;
  }
};

async function isDirectory(path: string) {
  return new Promise<boolean>((resolve, reject) => {
    fs.lstat(path, (err, stats) => {
      if (err) {
        reject(err);
      }
      resolve(stats.isDirectory());
    });
  });
}

export default function Main() {
  const {
    AppService,
    WindowsService,
    EditorCommandsService,
    ScenesService,
    CustomizationService,
  } = Services;
  const mainWindowEl = useRef<HTMLDivElement | null>(null);
  const mainMiddleEl = useRef<HTMLDivElement | null>(null);
  const windowResizeTimeout = useRef<number | null>(null);

  const [bulkLoadFinished, setBulkLoadFinished] = useState(false);
  const [i18nReady, seti18nReady] = useState(false);
  const [compactView, setCompactView] = useState(false);
  const [minEditorWidth, setMinEditorWidth] = useState(500);

  const uiReady = bulkLoadFinished && i18nReady;

  const { currentPage: page, params } = useRealmObject(Services.NavigationService.state);
  const { theme: realmTheme } = useRealmObject(Services.CustomizationService.state);

  const {
    errorAlert,
    applicationLoading,
    hideStyleBlockers,
    activeSceneId,
  } = useVuex(() => ({
    errorAlert: AppService.state.errorAlert,
    applicationLoading: AppService.state.loading,
    hideStyleBlockers: WindowsService.state.main.hideStyleBlockers,
    activeSceneId: ScenesService.views.activeSceneId,
  }));

  const showLoadingSpinner = useMemo(
    () => applicationLoading && page !== 'Onboarding' && page !== 'BrowseOverlays',
    [applicationLoading, page],
  );

  const theme = useMemo(() => {
    return !bulkLoadFinished ? loadedTheme() || 'night-theme' : realmTheme;
  }, [bulkLoadFinished, realmTheme]);

  const updateStyleBlockers = useCallback((val: boolean) => {
    WindowsService.actions.updateStyleBlockers('main', val);
  }, []);

  const onDropHandler = useCallback(
    async (event: React.DragEvent) => {
      if (page !== 'Studio') return;

      const fileList = event.dataTransfer?.files;

      if (!fileList || fileList.length < 1) return;

      const files: string[] = [];
      let fi = fileList.length;
      while (fi--) files.push(fileList.item(fi)!.path);

      const isDir = await isDirectory(files[0]).catch(err => {
        console.error('Error checking if drop is directory', err);
        return false;
      });

      if (files.length > 1 || isDir) {
        remote.dialog
          .showMessageBox(remote.getCurrentWindow(), {
            title: 'Streamlabs Desktop',
            message: $t('Are you sure you want to import multiple files?'),
            type: 'warning',
            buttons: [$t('Cancel'), $t('OK')],
          })
          .then(({ response }) => {
            if (!response) return;
            EditorCommandsService.actions.executeCommand('AddFilesCommand', activeSceneId, files);
          });
      } else {
        EditorCommandsService.actions.executeCommand('AddFilesCommand', activeSceneId, files);
      }
    },
    [activeSceneId, page],
  );

  const handleEditorWidth = useDebounce(500, (width: number) => {
    setMinEditorWidth(width);
  });

  function windowSizeHandler() {
    if (!hideStyleBlockers) {
      updateStyleBlockers(true);
    }
    const windowWidth = window.innerWidth;

    if (windowResizeTimeout.current) clearTimeout(windowResizeTimeout.current);

    windowResizeTimeout.current = window.setTimeout(() => {
      updateStyleBlockers(false);
    }, 200);
  }

  useEffect(() => {
    let unsubscribe: (() => void) | null = null;
    unsubscribe = StatefulService.store.subscribe((_, state) => {
      if (state.bulkLoadFinished) setBulkLoadFinished(true);
      if (state.i18nReady) seti18nReady(true);
      if (state.bulkLoadFinished && state.i18nReady && unsubscribe) {
        unsubscribe();
        unsubscribe = null;
      }
    });

    windowSizeHandler();

    return () => {
      if (unsubscribe) unsubscribe();
    };
  }, []);

  useEffect(() => {
    window.addEventListener('resize', windowSizeHandler);

    return () => {
      window.removeEventListener('resize', windowSizeHandler);
    };
  }, []);

  const oldTheme = useRef<TApplicationTheme | null>(null);
  useEffect(() => {
    if (!theme) return;
    if (oldTheme.current && oldTheme.current !== theme) antdThemes[oldTheme.current].unuse();
    antdThemes[theme].use();
    oldTheme.current = theme;
  }, [theme]);

  useEffect(() => {
    setCompactView(!!mainMiddleEl.current && mainMiddleEl.current.clientWidth < 1200);
  }, [uiReady, hideStyleBlockers]);

  if (!uiReady) return <div className={cx(styles.main, theme)} />;

  const Component: React.FunctionComponent<{
    className?: string;
    params: any;
    onTotalWidth: (width: number) => void;
  }> = (appPages[page as keyof typeof appPages] || appPages.Studio) as React.FunctionComponent<{
    className?: string;
    params: any;
    onTotalWidth: (width: number) => void;
  }>;

  return (
    <div
      className={cx(styles.main, theme, 'react')}
      id="mainWrapper"
      ref={mainWindowEl}
      onDrop={(ev: React.DragEvent) => onDropHandler(ev)}
    >
      <TitleBar windowId="main" className={cx({ [styles.titlebarError]: errorAlert })} />
      <div
        className={cx(styles.mainContents, {
          [styles.mainContentsOnboarding]: page === 'Onboarding',
        })}
      >
        {page !== 'Onboarding' && !showLoadingSpinner && (
          <div className={styles.sideNavContainer}>
            <SideNav />
          </div>
        )}
        <div
          className={cx(styles.mainMiddle, {
            [styles.mainMiddleCompact]: compactView,
            [styles.onboarding]: page === 'Onboarding',
          })}
          ref={mainMiddleEl}
        >
          {!showLoadingSpinner && (
            <div className={styles.mainPageContainer}>
              <Component
                params={params}
                onTotalWidth={(width: number) => handleEditorWidth(width)}
              />
            </div>
          )}
          {!applicationLoading && page !== 'Onboarding' && (
            <div style={{ display: 'flex', minWidth: '0px', gridRow: '2 / span 1' }}>
              <StudioFooter />
            </div>
          )}
        </div>
      </div>
      <Animation transitionName="ant-fade">
        {(!uiReady || showLoadingSpinner) && (
          <div className={cx(styles.mainLoading, { [styles.initialLoading]: !uiReady })}>
            <Loader />
          </div>
        )}
      </Animation>
    </div>
  );
}
