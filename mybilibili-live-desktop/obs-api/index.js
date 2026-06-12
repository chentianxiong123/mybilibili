'use strict';

exports.__esModule = true;
let obs;
try {
  obs = window['require']('obs-studio-node');
} catch (e) {
  console.error('[obs-api] obs-studio-node 加载失败,OBS 不可用:', e.message);
  obs = {};
}
if (!obs) obs = {};

/* Use for...in operator to perfectly mirror the osn module */
for (const entry in obs) {
  const url = new URL(window.location.href);

  if (url.searchParams.get('windowId') === 'worker') {
    exports[entry] = obs[entry];
  } else {
    exports[entry] = new Proxy(
      {},
      {
        get(target, property) {
          throw new Error(
            `Attempted to access OBS property ${property} outside of the worker process. OBS can only be accessed from the worker process.`,
          );
        },
      },
    );
  }
}
