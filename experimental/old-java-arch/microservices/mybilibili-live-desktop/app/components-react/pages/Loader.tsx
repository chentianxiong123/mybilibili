import React from 'react';
import cx from 'classnames';
import SvgContainer from 'components-react/shared/SvgContainer';

export default function Loader(p: { className?: string }) {
  return (
    <div className={cx('s-loader', p.className)}>
      <div className="s-loader__bg">
        <div className="s-loader__inner">
          <div className="s-loader__brand">mybilibili 直播工作台</div>
          <Spinner />
          <div className="s-loader__text">正在初始化直播工作台</div>
        </div>
      </div>
    </div>
  );
}

export function Spinner() {
  return (
    <div className="s-spinner s-spinner__overlay">
      <div className="s-bars">
        <SvgContainer src={spinnerSrc} className="s-spinner--large" />
      </div>
    </div>
  );
}

const spinnerSrc = `
<svg
version="1.1"
xmlns="http://www.w3.org/2000/svg"
xmlnsXlink="http://www.w3.org/1999/xlink"
viewBox="0 0 28 40"
>
  <path d="M0 0, l0 4, l0 -4" id="s-bar-y-path"></path>
  <rect width="4" height="40" x="0" y="0" ry="2" rx="2" class="s-spinner__bar">
    <animate
      attributeName="opacity"
      values=".24; .08; .24"
      begin="0s"
      dur="1.2s"
      repeatCount="indefinite"
    ></animate>
    <animate
      attributeName="height"
      values="40; 32; 40"
      begin="0s"
      dur="1.2s"
      repeatCount="indefinite"
    ></animate>
    <animateMotion begin="0s" dur="1.2s" repeatCount="indefinite">
      <mpath xlink:href="#s-bar-y-path"></mpath>
    </animateMotion>
  </rect>
  <rect width="4" height="40" x="12" y="0" ry="2" rx="2" class="s-spinner__bar">
    <animate attributeName="opacity" values=".24; .24; .24" begin="0s" dur="0.4s"></animate>
    <animate
      attributeName="opacity"
      values=".24; .08; .24"
      begin="0.4s"
      dur="1.2s"
      repeatCount="indefinite"
    ></animate>
    <animate
      attributeName="height"
      values="40; 32; 40"
      begin="0.4s"
      dur="1.2s"
      repeatCount="indefinite"
    ></animate>
    <animateMotion begin="0.4s" dur="1.2s" repeatCount="indefinite">
      <mpath xlink:href="#s-bar-y-path"></mpath>
    </animateMotion>
  </rect>
  <rect width="4" height="40" x="24" y="0" ry="2" rx="2" class="s-spinner__bar">
    <animate attributeName="opacity" values=".24; .24; .24" begin="0s" dur="0.8s"></animate>
    <animate
      attributeName="opacity"
      values=".24; .08; .24"
      begin="0.8s"
      dur="1.2s"
      repeatCount="indefinite"
    ></animate>
    <animate
      attributeName="height"
      values="40; 32; 40"
      begin="0.8s"
      dur="1.2s"
      repeatCount="indefinite"
    ></animate>
    <animateMotion begin="0.8s" dur="1.2s" repeatCount="indefinite">
      <mpath xlink:href="#s-bar-y-path"></mpath>
    </animateMotion>
  </rect>
</svg>`;
