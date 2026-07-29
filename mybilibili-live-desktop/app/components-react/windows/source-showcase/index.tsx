import React, { useMemo, useState } from 'react';
import { Input, Layout, Menu } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import cx from 'classnames';
import { ModalLayout } from 'components-react/shared/ModalLayout';
import Scrollable from 'components-react/shared/Scrollable';
import { Services } from 'components-react/service-provider';
import { SourceDisplayData, TSourceType } from 'services/sources';
import { $t } from 'services/i18n';
import styles from './SourceShowcase.m.less';

const { Content, Sider } = Layout;

type TSourceGroup = 'all' | 'capture' | 'av' | 'media';

export default function SourceShowcase() {
  const { SourcesService, WindowsService } = Services;
  const [activeTab, setActiveTab] = useState<TSourceGroup>('all');
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedSource, setSelectedSource] = useState<TSourceType | null>(null);

  const displayData = useMemo(() => SourceDisplayData(), []);
  const availableSources = useMemo(
    () =>
      SourcesService.getAvailableSourcesTypesList().filter(source => {
        const data = displayData[source.value];
        if (!data) return false;
        if (source.value === 'text_ft2_source') return false;
        return true;
      }),
    [],
  );

  const visibleSources = useMemo(() => {
    const term = searchTerm.trim().toLowerCase();

    return availableSources.filter(source => {
      const data = displayData[source.value];
      const matchesTab = activeTab === 'all' || data?.group === activeTab;
      const matchesSearch =
        !term ||
        source.description.toLowerCase().includes(term) ||
        data?.name?.toLowerCase().includes(term) ||
        data?.description?.toLowerCase().includes(term);

      return matchesTab && matchesSearch;
    });
  }, [activeTab, availableSources, displayData, searchTerm]);

  const inspectedSource = selectedSource || visibleSources[0]?.value || availableSources[0]?.value;
  const inspectedData = inspectedSource ? displayData[inspectedSource] : null;

  function addSource() {
    if (!inspectedSource) return;
    WindowsService.actions.closeChildWindow();
    SourcesService.actions.showAddSource(inspectedSource);
  }

  return (
    <ModalLayout onOk={addSource} okText={$t('Add Source')} bodyStyle={{ padding: 0 }}>
      <Layout style={{ height: '100%' }}>
        <Content style={{ minWidth: 0 }}>
          <div className={styles.header}>
            <Menu
              onClick={e => setActiveTab(e.key as TSourceGroup)}
              selectedKeys={[activeTab]}
              mode="horizontal"
              style={{ borderBottom: 0, flex: 1 }}
            >
              <Menu.Item key="all">{$t('All Sources')}</Menu.Item>
              <Menu.Item key="capture">{$t('Capture Sources')}</Menu.Item>
              <Menu.Item key="av">{$t('Video and Audio')}</Menu.Item>
              <Menu.Item key="media">{$t('Media')}</Menu.Item>
            </Menu>
            <Input
              type="search"
              className={styles.search}
              allowClear
              placeholder={$t('Search...')}
              prefix={<SearchOutlined />}
              onChange={ev => setSearchTerm(ev.target.value)}
              value={searchTerm}
            />
          </div>
          <Scrollable style={{ height: 'calc(100% - 46px)', padding: 16 }}>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gap: 8 }}>
              {visibleSources.map(source => {
                const data = displayData[source.value];
                const active = inspectedSource === source.value;

                return (
                  <button
                    key={source.value}
                    type="button"
                    className={cx(styles.sourceTag, { [styles.active]: active })}
                    onClick={() => setSelectedSource(source.value)}
                    onDoubleClick={() => {
                      setSelectedSource(source.value);
                      SourcesService.actions.showAddSource(source.value);
                    }}
                  >
                    <span className={styles.nameRow}>
                      <i className={data?.icon || 'icon-add'} />
                      <strong>{data?.name || source.description}</strong>
                    </span>
                    <span>{data?.shortDesc || source.description}</span>
                  </button>
                );
              })}
            </div>
          </Scrollable>
        </Content>
        <Sider width={300} style={{ height: '100%' }} collapsed={!inspectedData} collapsedWidth={0}>
          <div className={styles.preview}>
            <Scrollable style={{ height: '100%' }}>
              <h2 style={{ marginTop: 0 }}>{inspectedData?.name}</h2>
              <div>{inspectedData?.description}</div>
              {inspectedData?.supportList?.length > 0 && (
                <>
                  <div className={styles.supportHeader}>{$t('Supports:')}</div>
                  <ul style={{ fontSize: 13 }}>
                    {inspectedData.supportList.map(support => (
                      <li key={support}>{support}</li>
                    ))}
                  </ul>
                </>
              )}
            </Scrollable>
          </div>
        </Sider>
      </Layout>
    </ModalLayout>
  );
}
