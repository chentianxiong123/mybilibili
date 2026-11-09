const appStartTime = Date.now();
let lastEventTime = 0;

////////////////////////////////////////////////////////////////////////////////
// Set Up Environment Variables
////////////////////////////////////////////////////////////////////////////////
const pjson = require('./package.json');
if (pjson.env === 'production') {
  process.env.NODE_ENV = 'production';
}
if (pjson.name === 'slobs-client-preview') {
  process.env.SLOBS_PREVIEW = true;
}
if (pjson.name === 'slobs-client-ipc') {
  process.env.SLOBS_IPC = true;
}
process.env.SLOBS_VERSION = pjson.version;
const isDevMode = process.env.NODE_ENV !== 'production';

////////////////////////////////////////////////////////////////////////////////
// Modules and other Requires
////////////////////////////////////////////////////////////////////////////////
const {
  app,
  BrowserWindow,
  ipcMain,
  session,
  dialog,
  webContents,
  desktopCapturer,
  MessageChannelMain,
  protocol,
} = require('electron');
const path = require('path');
const remote = require('@electron/remote/main');
const fs = require('fs');

// We use a special cache directory for running tests
if (process.env.SLOBS_CACHE_DIR) {
  app.setPath('appData', process.env.SLOBS_CACHE_DIR);
}

app.setPath('userData', path.join(app.getPath('appData'), 'mybilibili-live-desktop'));

// 注册 slbundle:// 自定义协议:privileged 必须在 app ready 之前注册
protocol.registerSchemesAsPrivileged([
  { scheme: 'slbundle', privileges: { standard: true, secure: true, supportFetchAPI: true, corsEnabled: true, stream: true } }
]);

if (process.argv.includes('--clearCacheDir')) {
  try {
    // This could block for a while, but should ensure that the crash handler
    // is no longer able to interfere with cache removal.
    fs.rmSync(app.getPath('userData'), {
      force: true,
      recursive: true,
      maxRetries: 5,
      retryDelay: 500,
    });
  } catch (e) {}
}

// This ensures that only one copy of our app can run at once.
const gotTheLock = app.requestSingleInstanceLock();

if (!gotTheLock) {
  app.quit();
  return;
}

const uuid = require('uuid/v4');
const windowStateKeeper = require('electron-window-state');

app.commandLine.appendSwitch('force-ui-direction', 'ltr');
app.commandLine.appendSwitch('ignore-connections-limit', 'localhost,127.0.0.1');

process.env.IPC_UUID = `mybilibili-live-${uuid()}`;

////////////////////////////////////////////////////////////////////////////////
// Main Program
////////////////////////////////////////////////////////////////////////////////

// Windows
let workerWindow;
let mainWindow;
let childWindow;

const util = require('util');
const logFile = path.join(app.getPath('userData'), 'app.log');
const maxLogBytes = 131072;

// Truncate the log file if it is too long
if (fs.existsSync(logFile) && fs.statSync(logFile).size > maxLogBytes) {
  const content = fs.readFileSync(logFile);
  fs.writeFileSync(logFile, '[LOG TRUNCATED]\n');
  fs.writeFileSync(logFile, content.slice(content.length - maxLogBytes), { flag: 'a' });
}

ipcMain.on('logmsg', (e, msg) => {
  if (msg.level === 'error' && mainWindow && process.env.NODE_ENV !== 'production') {
    mainWindow.send('unhandledErrorState');
  }

  logFromRemote(msg.level, msg.sender, msg.message);
});

// app.ts 调 electron.ipcRenderer.sendSync('getBundleNames', bundles),
// 原版用真实 bundle 名替换 stack trace。fork 时没注册,导致渲染卡死
ipcMain.on('getBundleNames', (e, bundles) => {
  // 直接返回 file:// URL 形式,renderer 拿来当 filename 用即可
  const result = {};
  bundles.forEach(name => {
    result[name] = `file://${__dirname}/bundles/${name}`;
  });
  e.returnValue = result;
});

function logFromRemote(level, sender, msg) {
  msg.split('\n').forEach(line => {
    writeLogLine(`[${new Date().toISOString()}] [${level}] [${sender}] - ${line}`);
  });
}

function serializeConsoleArg(arg) {
  if (typeof arg === 'string') return arg;
  if (arg instanceof Error) return arg.stack || arg.message;
  return util.inspect(arg);
}

function logMainConsole(level, args) {
  if (!process.env.SLOBS_DISABLE_MAIN_LOGGING) {
    const serialized = args.map(serializeConsoleArg).join(' ');

    logFromRemote(level, 'electron-main', serialized);
  }
}

const consoleLog = console.log;
console.log = (...args) => {
  logMainConsole('info', args);
};

console.warn = (...args) => {
  logMainConsole('warn', args);
};

console.error = (...args) => {
  logMainConsole('error', args);
};

const lineBuffer = [];

function writeLogLine(line) {
  // Also print to stdout
  consoleLog(line);

  lineBuffer.push(`${line}\n`);
  flushNextLine();
}

let writeInProgress = false;

function flushNextLine() {
  if (lineBuffer.length === 0) return;
  if (writeInProgress) return;

  const nextLine = lineBuffer.shift();

  writeInProgress = true;

  fs.writeFile(logFile, nextLine, { flag: 'a' }, e => {
    writeInProgress = false;

    if (e) {
      consoleLog('Error writing to log file', e);
      return;
    }

    flushNextLine();
  });
}

const os = require('os');
const cpus = os.cpus();

// Source: https://stackoverflow.com/questions/10420352/converting-file-size-in-bytes-to-human-readable-string/10420404
function humanFileSize(bytes, si) {
  const thresh = si ? 1000 : 1024;
  if (Math.abs(bytes) < thresh) {
    return bytes + ' B';
  }
  const units = si
    ? ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
    : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
  let u = -1;
  do {
    bytes /= thresh;
    ++u;
  } while (Math.abs(bytes) >= thresh && u < units.length - 1);
  return bytes.toFixed(1) + ' ' + units[u];
}

console.log('=================================');
console.log('mybilibili Live Desktop');
console.log(`Version: ${process.env.SLOBS_VERSION}`);
console.log(`OS: ${os.platform()} ${os.release()}`);
console.log(`Arch: ${process.arch}`);
console.log(`CPU: ${cpus[0].model}`);
console.log(`Cores: ${cpus.length}`);
console.log(`Memory: ${humanFileSize(os.totalmem(), false)}`);
console.log(`Free: ${humanFileSize(os.freemem(), false)}`);
console.log('=================================');

app.on('ready', () => {
  /* 注册 slbundle:// 协议(fork 后缺失,导致 renderer.js 加载失败)
     registerSchemesAsPrivileged 已在 ready 之前调用,这里只挂 handler */
  if (typeof protocol.handle === 'function') {
    protocol.handle('slbundle', (request) => {
      try {
        const url = new URL(request.url);
        // slbundle://bundles/renderer.js -> bundles/renderer.js
        // slbundle://vendor/xxx.css  -> vendor/xxx.css (可能缺失,返回空)
        const filePath = path.join(__dirname, url.hostname + url.pathname);
        if (fs.existsSync(filePath)) {
          console.log('[slbundle] 请求:', request.url, '-> 解析路径:', filePath);
          // 必须返回 Response 并设置正确 MIME 类型,否则 Chromium 拒绝执行
          const ext = path.extname(filePath).toLowerCase();
          const mimeTypes = { '.js': 'application/javascript', '.css': 'text/css', '.html': 'text/html', '.json': 'application/json' };
          const contentType = mimeTypes[ext] || 'application/octet-stream';
          const content = fs.readFileSync(filePath);
          return new Response(content, {
            status: 200,
            headers: {
              'Content-Type': `${contentType}; charset=utf-8`,
              'Access-Control-Allow-Origin': '*',
              'Cache-Control': 'no-store',
            },
          });
        } else {
          // vendor/ 目录在 fork 中缺失,返回空响应避免 CSP 报错
          console.log('[slbundle] 文件不存在,返回空:', filePath);
          const ext = path.extname(filePath).toLowerCase();
          if (ext === '.css' || ext === '.js') {
            return new Response('', { status: 200, headers: { 'Content-Type': 'text/' + (ext === '.js' ? 'javascript' : 'css') } });
          }
          return new Response('not found', { status: 404 });
        }
      } catch (e) {
        console.error('[slbundle] 解析失败:', request.url, e);
        return new Response('not found', { status: 404 });
      }
    });
    console.log('[ok] slbundle:// 协议已注册 (handle)');
  } else {
    protocol.registerFileProtocol('slbundle', (request, callback) => {
      const url = new URL(request.url);
      const filePath = path.join(__dirname, url.hostname + url.pathname);
      callback({ path: filePath });
    });
    console.log('[ok] slbundle:// 协议已注册 (legacy)');
  }

  /* Load React DevTools in dev mode */
  if (process.env.NODE_ENV === 'development') {
    const reactDevToolsPath = path.join(__dirname, 'vendor', 'react-devtools');
    session.defaultSession
      .loadExtension(reactDevToolsPath, { allowFileAccess: true })
      .then(() => console.log('Installed React DevTools'))
      .catch(err => console.log('Error installing React DevTools', err));
  }

  // Detect when running from an unwritable location like a DMG image (will break updater)
  if (process.platform === 'darwin') {
    try {
      fs.accessSync(app.getPath('exe'), fs.constants.W_OK);
    } catch (e) {
      // This error code indicates a read only file system
      if (e.code === 'EROFS') {
        dialog.showErrorBox(
          'mybilibili Live Desktop',
          '请从“应用程序”文件夹启动 mybilibili Live Desktop，不能直接从磁盘镜像运行。',
        );
        app.exit();
      }
    }
  }

  // network logging is disabled by default
  if (!process.argv.includes('--network-logging')) return;

  // ignore fs requests
  const filter = { urls: ['https://*', 'http://*'] };

  session.defaultSession.webRequest.onBeforeRequest(filter, (details, callback) => {
    console.log('HTTP REQUEST', details.method, details.url);
    callback(details);
  });

  session.defaultSession.webRequest.onErrorOccurred(filter, details => {
    console.log('HTTP REQUEST FAILED', details.method, details.url);
  });

  session.defaultSession.webRequest.onCompleted(filter, details => {
    console.log('HTTP REQUEST COMPLETED', details.method, details.url, details.statusCode);
  });
});

// Somewhat annoyingly, this is needed so that the main window
// can differentiate between a user closing it vs the app
// closing the windows before exit.
let allowMainWindowClose = false;
let shutdownStarted = false;
let appShutdownTimeout;

global.indexUrl = `file://${__dirname}/index.html`;

function openDevTools() {
  childWindow.webContents.openDevTools({ mode: 'detach' });
  mainWindow.webContents.openDevTools({ mode: 'detach' });
  workerWindow.webContents.openDevTools({ mode: 'detach' });
}

function attachWebContentsDiagnostics(name, win) {
  win.webContents.on('console-message', (event, levelOrDetails, message, line, sourceId) => {
    if (levelOrDetails && typeof levelOrDetails === 'object') {
      const details = levelOrDetails;
      const detailsMessage =
        details.message ||
        (typeof details.toString === 'function' ? details.toString() : JSON.stringify(details));
      console.log(
        `[renderer:${name}] ${details.level || ''} ${details.sourceId || ''}:${
          details.lineNumber || details.line || ''
        } - ${detailsMessage}`,
      );
      return;
    }

    console.log(`[renderer:${name}] ${sourceId || ''}:${line || ''} - ${message || ''}`);
  });

  win.webContents.on('did-fail-load', (event, errorCode, errorDescription, validatedURL) => {
    console.error(
      `[renderer:${name}] did-fail-load ${errorCode} ${errorDescription} ${validatedURL}`,
    );
  });

  win.webContents.on('render-process-gone', (event, details) => {
    console.error(`[renderer:${name}] render-process-gone ${JSON.stringify(details)}`);
  });

  win.webContents.on('did-finish-load', () => {
    console.log(`[renderer:${name}] did-finish-load ${win.webContents.getURL()}`);
  });
}

// TODO: Clean this up
// These windows are waiting for services to be ready
const waitingVuexStores = [];
let workerInitFinished = false;

async function startApp() {
  ipcMain.on('register-in-crash-handler', () => {});
  ipcMain.on('unregister-in-crash-handler', () => {});
  ipcMain.on('startupError', (e, msg) => {
    console.error('[DIAG-STARTUP-ERROR]', msg);
  });

  remote.initialize();

  workerWindow = new BrowserWindow({
    show: false,
    webPreferences: { nodeIntegration: true, contextIsolation: false },
  });

  remote.enable(workerWindow.webContents);
  attachWebContentsDiagnostics('worker', workerWindow);

  // setTimeout(() => {
  workerWindow.loadURL(`${global.indexUrl}?windowId=worker`);
  // }, 10 * 1000);

  if (process.env.SLOBS_PRODUCTION_DEBUG) {
    workerWindow.webContents.once('dom-ready', () => {
      workerWindow.webContents.openDevTools({ mode: 'detach' });
    });
  }

  // All renderers should use ipcRenderer.sendTo to send to communicate with
  // the worker.  This still gets proxied via the main process, but eventually
  // we will refactor this to not use electron IPC, which will make it much
  // more efficient.
  ipcMain.on('getWorkerWindowId', event => {
    if (workerWindow.isDestroyed()) {
      // prevent potential race-condition issues on app close
      // https://github.com/streamlabs/desktop/pull/4239
      return;
    }
    event.returnValue = workerWindow.webContents.id;
  });

  const mainWindowState = windowStateKeeper({
    defaultWidth: 1600,
    defaultHeight: 1000,
  });

  mainWindow = new BrowserWindow({
    minWidth: 800,
    minHeight: 600,
    width: mainWindowState.width,
    height: mainWindowState.height,
    x: mainWindowState.isMaximized ? mainWindowState.displayBounds.x : mainWindowState.x,
    y: mainWindowState.isMaximized ? mainWindowState.displayBounds.y : mainWindowState.y,
    show: false,
    frame: false,
    titleBarStyle: 'hidden',
    title: 'mybilibili Live Desktop',
    backgroundColor: '#17242D',
    webPreferences: {
      nodeIntegration: true,
      webviewTag: true,
      contextIsolation: false,
    },
  });

  remote.enable(mainWindow.webContents);
  attachWebContentsDiagnostics('main', mainWindow);

  // setTimeout(() => {
  mainWindow.loadURL(`${global.indexUrl}?windowId=main`);
  // }, 5 * 1000)

  // 诊断1:1 秒后确认 webpack boot 是否完成
  setTimeout(() => {
    workerWindow.webContents.executeJavaScript('JSON.stringify({t:1,w:location.search,d:document.title})').then(function(r){
      console.log('[DIAG-W1]', r);
    }).catch(function(e){ console.error('[DIAG-W1-ERR]', e.message); });
    mainWindow.webContents.executeJavaScript('JSON.stringify({t:1,w:location.search,d:document.title})').then(function(r){
      console.log('[DIAG-M1]', r);
    }).catch(function(e){ console.error('[DIAG-M1-ERR]', e.message); });
  }, 1000);

  // 诊断2:3 秒后问 worker 状态
  setTimeout(() => {
    workerWindow.webContents.executeJavaScript(`
      JSON.stringify({
        win: window.location.search,
        hasObs: typeof window.obs !== 'undefined',
        hasStore: typeof window.__vuex_store__ !== 'undefined',
        // webpack 5 挂在 window 上的 chunk 数组
        hasChunks: typeof webpackChunkmybilibili_live_desktop !== 'undefined',
        chunksLen: typeof webpackChunkmybilibili_live_desktop !== 'undefined' ? webpackChunkmybilibili_live_desktop.length : -1,
        // window 整体状态
        hasVue: typeof Vue !== 'undefined',
        hasReact: typeof React !== 'undefined',
        hasApp: !!document.getElementById('app'),
        appHTML: (document.getElementById('app') || {}).innerHTML ? document.getElementById('app').innerHTML.substring(0,300) : '',
        OBS_API_result: typeof window.__OBS_API_result !== 'undefined' ? window.__OBS_API_result : 'not_set',
        title: document.title,
        keys: Object.keys(window).length,
        sampleKeys: Object.keys(window).filter(k => k.startsWith('__') || k.startsWith('webpack') || k.startsWith('Vue') || k.startsWith('React') || k.startsWith('obs')).join(',')
      })
    `).then(function(r) {
      try { var s = JSON.parse(r); console.log('[DIAG-WORKER]', JSON.stringify(s)); }
      catch(e) { console.log('[DIAG-WORKER]', r); }
    }).catch(function(e) { console.error('[DIAG-WORKER-ERR]', e.message); });
  }, 3000);

  // 诊断:4 秒后问 main window 状态
  setTimeout(() => {
    mainWindow.webContents.executeJavaScript(`
      JSON.stringify({
        win: window.location.search,
        hasStore: typeof window.__vuex_store__ !== 'undefined',
        bulkLoadFinished: window.__vuex_store__ ? window.__vuex_store__.state.bulkLoadFinished : 'no_store',
        i18nReady: window.__vuex_store__ ? window.__vuex_store__.state.i18nReady : 'no_store',
        title: document.title,
        appHTML: (document.getElementById('app') || document.body || {}).innerHTML ? (document.getElementById('app') || document.body).innerHTML.substring(0,300) : '',
        bodyText: document.body ? document.body.innerText.substring(0,300) : '',
        registeredStores: typeof window.__vuex_store__ !== 'undefined' ? 'store_exists' : 'no_store',
        keys: Object.keys(window).length
      })
    `).then(function(r) {
      try { var s = JSON.parse(r); console.log('[DIAG-MAIN]', JSON.stringify(s)); }
      catch(e) { console.log('[DIAG-MAIN]', r); }
    }).catch(function(e) { console.error('[DIAG-MAIN-ERR]', e.message); });
  }, 4000);

  mainWindowState.manage(mainWindow);

  mainWindow.removeMenu();

  // 显示窗口 - 修复 fork 后窗口未自动显示的 bug
  mainWindow.once('ready-to-show', () => {
    mainWindow.show();
  });
  // 兜底:加载完后 3 秒强制显示,避免 ready-to-show 不触发
  setTimeout(() => {
    if (mainWindow && !mainWindow.isDestroyed() && !mainWindow.isVisible()) {
      mainWindow.show();
    }
  }, 3000);

  mainWindow.on('close', e => {
    if (!shutdownStarted) {
      shutdownStarted = true;
      workerWindow.send('shutdown');

      // We give the worker window 10 seconds to acknowledge a request
      // to shut down.  Otherwise, we just close it.
      appShutdownTimeout = setTimeout(() => {
        allowMainWindowClose = true;
        if (!mainWindow.isDestroyed()) mainWindow.close();
        if (!workerWindow.isDestroyed()) workerWindow.close();
      }, 10 * 1000);
    }

    if (!allowMainWindowClose) e.preventDefault();
  });

  // prevent worker window to be closed before other windows
  // we need it to properly handle App.stop() in tests
  // since it tries to close all windows
  workerWindow.on('close', e => {
    if (!shutdownStarted) {
      e.preventDefault();
      mainWindow.close();
    }
  });

  // This needs to be explicitly handled on Mac
  app.on('before-quit', e => {
    if (!shutdownStarted) {
      e.preventDefault();
      mainWindow.close();
    }
  });

  ipcMain.on('acknowledgeShutdown', () => {
    if (appShutdownTimeout) clearTimeout(appShutdownTimeout);
  });

  ipcMain.on('shutdownComplete', () => {
    allowMainWindowClose = true;
    mainWindow.close();
    workerWindow.close();
  });

  workerWindow.on('closed', () => {
    session.defaultSession.flushStorageData();
    session.defaultSession.cookies.flushStore().then(() => app.quit());
  });

  // Pre-initialize the child window
  childWindow = new BrowserWindow({
    show: false,
    frame: false,
    fullscreenable: false,
    titleBarStyle: 'hidden',
    backgroundColor: '#17242D',
    webPreferences: {
      nodeIntegration: true,
      backgroundThrottling: false,
      contextIsolation: false,
    },
  });

  remote.enable(childWindow.webContents);
  attachWebContentsDiagnostics('child', childWindow);

  childWindow.removeMenu();

  childWindow.loadURL(`${global.indexUrl}?windowId=child`);

  if (process.env.SLOBS_PRODUCTION_DEBUG) {
    childWindow.webContents.once('dom-ready', () => {
      childWindow.webContents.openDevTools({ mode: 'detach' });
    });
  }

  // The child window is never closed, it just hides in the
  // background until it is needed.
  childWindow.on('close', e => {
    if (!shutdownStarted) {
      childWindow.send('closeWindow');

      // Prevent the window from actually closing
      e.preventDefault();
    }
  });

  // simple messaging system for services between windows
  // WARNING! renderer windows use synchronous requests and will be frozen
  // until the worker window's asynchronous response
  const requests = {};

  function sendRequest(request, event = null, async = false) {
    if (workerWindow.isDestroyed()) {
      console.log('Tried to send request but worker window was missing...');
      return;
    }
    workerWindow.webContents.send('services-request', request);
    if (!event) return;
    requests[request.id] = Object.assign({}, request, { event, async });
  }

  // use this function to call some service method from the main process
  function callService(resource, method, ...args) {
    sendRequest({
      jsonrpc: '2.0',
      method,
      params: {
        resource,
        args,
      },
    });
  }

  ipcMain.on('AppInitFinished', () => {
    console.log('[DIAG] AppInitFinished 收到! worker 初始化完成');
    workerInitFinished = true;

    waitingVuexStores.forEach(winId => {
      console.log('[DIAG] 发送 initFinished 给窗口', winId);
      BrowserWindow.fromId(winId).send('initFinished');
    });

    waitingVuexStores.forEach(windowId => {
      console.log('[DIAG] 通知 worker 发送 vuex 状态给窗口', windowId);
      workerWindow.webContents.send('vuex-sendState', windowId);
    });
    console.log('[DIAG] AppInitFinished 处理完毕');
  });

  ipcMain.on('vuex-sendState', () => {
    console.log('[DIAG] worker 回复状态同步');
  });

  ipcMain.on('vuex-mutation', (event, mutation) => {
    // 只打印关键 mutations
    const key = mutation && mutation.payload && mutation.payload[0];
    if (key && key.includes('bulkLoad') || key && key.includes('i18n')) {
      console.log('[DIAG] vuex-mutation:', key, '=', mutation.payload[1]);
    }
  });

  // 监控 shutdownComplete 和 unhandledErrorState
  ipcMain.on('shutdownComplete', () => {
    console.log('[DIAG] shutdownComplete! worker 进程关闭');
  });

  ipcMain.on('services-request', (event, payload) => {
    sendRequest(payload, event);
  });

  ipcMain.on('services-request-async', (event, payload) => {
    sendRequest(payload, event, true);
  });

  ipcMain.on('services-response', (event, response) => {
    if (!requests[response.id]) return;

    if (requests[response.id].async) {
      requests[response.id].event.reply('services-response-async', response);
    } else {
      requests[response.id].event.returnValue = response;
    }
    delete requests[response.id];
  });

  ipcMain.on('services-message', (event, payload) => {
    const windows = BrowserWindow.getAllWindows();
    windows.forEach(window => {
      if (window.id === workerWindow.id || window.isDestroyed()) return;
      window.webContents.send('services-message', payload);
    });
  });

  if (isDevMode) {
    // Vue dev tools appears to cause strange non-deterministic
    // interference with certain NodeJS APIs, expecially asynchronous
    // IO from the renderer process.  Enable at your own risk.
    // const devtoolsInstaller = require('electron-devtools-installer');
    // devtoolsInstaller.default(devtoolsInstaller.VUEJS_DEVTOOLS);
    // setTimeout(() => {
    //   openDevTools();
    // }, 10 * 1000);
  }
}

const haDisableFile = path.join(app.getPath('userData'), 'HADisable');
if (fs.existsSync(haDisableFile)) app.disableHardwareAcceleration();

app.setAsDefaultProtocolClient('mybilibili-live');

app.on('second-instance', (event, argv, cwd) => {
  // Check for protocol links in the argv of the other process
  argv.forEach(arg => {
    if (arg.match(/^mybilibili-live:\/\//)) {
      workerWindow.send('protocolLink', arg);
    }
  });

  // Someone tried to run a second instance, we should focus our window.
  if (mainWindow && !mainWindow.isDestroyed()) {
    if (mainWindow.isMinimized()) {
      mainWindow.restore();
    }

    mainWindow.focus();
  } else if (!shutdownStarted) {
    // This instance is a zombie and we should shut down.
    app.exit();
  }
});

let protocolLinkReady = false;
let pendingLink;

// For mac os, this event will fire when a protocol link is triggered
app.on('open-url', (e, url) => {
  if (protocolLinkReady) {
    workerWindow.send('protocolLink', url);
  } else {
    pendingLink = url;
  }
});

ipcMain.on('protocolLinkReady', () => {
  protocolLinkReady = true;
  if (pendingLink) workerWindow.send('protocolLink', pendingLink);
});

app.on('ready', () => {
  startApp();
});

ipcMain.on('openDevTools', () => {
  openDevTools();
});

ipcMain.on('window-closeChildWindow', event => {
  // never close the child window, hide it instead
  if (!childWindow.isDestroyed()) childWindow.hide();
});

ipcMain.on('window-focusMain', () => {
  if (!mainWindow.isDestroyed()) mainWindow.focus();
});

// The main process acts as a hub for various windows
// syncing their vuex stores.
const registeredStores = {};

ipcMain.on('vuex-register', event => {
  const win = BrowserWindow.fromWebContents(event.sender);
  const windowId = win.id;

  // Register can be received multiple times if the window is
  // refreshed.  We only want to register it once.
  if (!registeredStores[windowId]) {
    registeredStores[windowId] = win;
    console.log('Registered vuex stores: ', Object.keys(registeredStores));

    // Make sure we unregister is when it is closed
    win.on('closed', () => {
      delete registeredStores[windowId];
      console.log('Registered vuex stores: ', Object.keys(registeredStores));
    });
  }

  if (windowId !== workerWindow.id) {
    // Tell the worker window to send its current store state
    // to the newly registered window

    if (workerInitFinished) {
      win.send('initFinished');
      workerWindow.webContents.send('vuex-sendState', windowId);
    } else {
      waitingVuexStores.push(windowId);
    }
  }
});

// Proxy vuex-mutation events to all other subscribed windows
ipcMain.on('vuex-mutation', (event, mutation) => {
  const senderWindow = BrowserWindow.fromWebContents(event.sender);

  if (senderWindow && !senderWindow.isDestroyed()) {
    const windowId = senderWindow.id;

    Object.keys(registeredStores)
      .filter(id => id !== windowId.toString())
      .forEach(id => {
        const win = registeredStores[id];
        if (!win.isDestroyed()) win.webContents.send('vuex-mutation', mutation);
      });
  }
});

ipcMain.on('restartApp', () => {
  app.relaunch();
  // Closing the main window starts the shut down sequence
  mainWindow.close();
});

/* The following 3 methods need to live in the main process
    because events bound using the remote module are not
    executed synchronously and therefore default actions
    cannot be prevented. */
ipcMain.on('webContents-preventNavigation', (e, id) => {
  const contents = webContents.fromId(id);

  if (contents.isDestroyed()) return;

  contents.on('will-navigate', e => {
    e.preventDefault();
  });
});

ipcMain.on('webContents-bindYTChat', (e, id) => {
  const contents = webContents.fromId(id);

  if (contents.isDestroyed()) return;

  contents.on('will-navigate', (e, targetUrl) => {
    const url = require('url');
    const parsed = url.parse(targetUrl);

    if (parsed.hostname === 'accounts.google.com') {
      e.preventDefault();
    }
  });
});

ipcMain.on('webContents-enableRemote', (e, id) => {
  const contents = webContents.fromId(id);

  if (contents.isDestroyed()) return;

  remote.enable(contents);

  // Needed otherwise the renderer will lock up
  e.returnValue = null;
});

ipcMain.on('getMainWindowWebContentsId', e => {
  e.returnValue = mainWindow.webContents.id;
});

ipcMain.on('requestPerformanceStats', e => {
  const stats = app.getAppMetrics();
  e.sender.send('performanceStatsResponse', stats);
});

ipcMain.on('showErrorAlert', () => {
  if (!mainWindow.isDestroyed()) {
    // main window may be destroyed on shutdown
    mainWindow.send('showErrorAlert');
  }
});

ipcMain.on('gameOverlayPaintCallback', (e, { contentsId, overlayId }) => {
  return;
});

ipcMain.on('getWindowIds', e => {
  e.returnValue = {
    worker: workerWindow.id,
    main: mainWindow.id,
    child: childWindow.id,
  };
});

ipcMain.on('getAppStartTime', e => {
  e.returnValue = appStartTime;
});

ipcMain.on('measure-time', (e, msg, time) => {
  measure(msg, time);
});

// Measure time between events
function measure(msg, time) {
  if (!time) time = Date.now();
  const delta = lastEventTime ? time - lastEventTime : 0;
  lastEventTime = time;
  if (delta > 2000) console.log('------------------');
  console.log(msg, delta + 'ms');
}

ipcMain.handle('DESKTOP_CAPTURER_GET_SOURCES', (event, opts) => desktopCapturer.getSources(opts));

// Message channel handling
const channels = {};

ipcMain.handle('create-message-channel', () => {
  const id = uuid();
  channels[id] = new MessageChannelMain();
  return id;
});

ipcMain.on('request-message-channel-in', (e, id) => {
  e.senderFrame.postMessage(`port-${id}`, null, [channels[id].port1]);
});

ipcMain.on('request-message-channel-out', (e, id) => {
  e.senderFrame.postMessage(`port-${id}`, null, [channels[id].port2]);
});
