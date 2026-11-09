import React from 'react';
import { Col } from 'antd';
import cx from 'classnames';
import { SourceDisplayData } from 'services/sources';
import styles from './SourceShowcase.m.less';

export default function SourceTag(p: {
  name?: string;
  type: string;
  essential?: boolean;
  excludeWrap?: boolean;
  hideShortDescription?: boolean;
  active?: boolean;
  onClick?: () => void;
}) {
  const displayData = SourceDisplayData()[p.type];

  return (
    <Col span={8}>
      <div
        className={cx('source-tag', styles.sourceTag, {
          [styles.active]: p.active,
          [styles.essential]: p.essential,
          [styles.excludeWrap]: p.excludeWrap,
        })}
        onClick={p.onClick}
        data-name={displayData?.name || p.name}
      >
        <div className={styles.nameRow}>
          <div className={styles.iconWrapper}>
            {displayData?.icon && <i className={displayData.icon} />}
          </div>
          <div className={styles.displayName}>{displayData?.name || p.name}</div>
        </div>
        {displayData?.shortDesc && !p.hideShortDescription && (
          <div style={{ opacity: 0.5 }}>{displayData.shortDesc}</div>
        )}
      </div>
    </Col>
  );
}
