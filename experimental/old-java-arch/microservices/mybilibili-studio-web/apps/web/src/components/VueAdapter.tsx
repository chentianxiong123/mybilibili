import React, { useEffect, useRef } from "react";
import { createApp, type App as VueApp } from "vue";

interface VueAdapterProps {
  component: any;
  props?: Record<string, any>;
  className?: string;
}

export const VueAdapter: React.FC<VueAdapterProps> = ({
  component,
  props = {},
  className = "",
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const vueAppRef = useRef<VueApp | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    // 创建 Vue 实例并挂载
    // 为了能够动态更新 props，我们可以包裹一个高阶响应式组件，或者直接通过 reactive 对象传值
    // 这里使用最直观的写法：利用 reactive 对象传递 props，更新 props 时直接修改该对象
    const app = createApp(component, props);
    vueAppRef.current = app;
    app.mount(containerRef.current);

    return () => {
      if (vueAppRef.current) {
        vueAppRef.current.unmount();
        vueAppRef.current = null;
      }
    };
  }, [component]);

  // 当 React props 改变时，由于 Vue App 已经 mount，不能直接重传 props
  // 简易且安全的方案：如果 props 改变，我们卸载并重新挂载，或者在 Vue 组件内部使用全局 Store / 事件
  // 对于 Toolbar 这类无重度内部状态的组件，若 props 改变较少，在 props 变化时重新 mount 是最简单可靠的桥接方式
  const prevPropsStr = JSON.stringify(props);
  useEffect(() => {
    if (vueAppRef.current && containerRef.current) {
      // 仅在 props 发生实质变化时重新挂载
      vueAppRef.current.unmount();
      const app = createApp(component, props);
      vueAppRef.current = app;
      app.mount(containerRef.current);
    }
  }, [prevPropsStr]);

  return <div ref={containerRef} className={className} />;
};

export default VueAdapter;
