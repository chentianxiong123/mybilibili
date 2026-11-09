'use strict';

exports.__esModule = true;
const obs = window['require']('obs-studio-node');

const runtimeEnums = {
  ESourceFlags: {
    Unbuffered: 1,
    ForceMono: 2,
  },
  EMonitoringType: {
    None: 0,
    MonitoringOnly: 1,
    MonitoringAndOutput: 2,
  },
  EOrderMovement: {
    Up: 0,
    Down: 1,
    Top: 2,
    Bottom: 3,
  },
  EDeinterlaceFieldOrder: {
    Top: 0,
    Bottom: 1,
  },
  EVideoCodes: {
    Success: 0,
    Fail: -1,
    NotSupported: -2,
    InvalidParam: -3,
    CurrentlyActive: -4,
    ModuleNotFound: -5,
  },
  EHotkeyObjectType: {
    Frontend: 0,
    Source: 1,
    Output: 2,
    Encoder: 3,
    Service: 4,
  },
  EDeinterlaceMode: {
    Disable: 0,
    Discard: 1,
    Retro: 2,
    Blend: 3,
    Blend2X: 4,
    Linear: 5,
    Linear2X: 6,
    Yadif: 7,
    Yadif2X: 8,
  },
  EBlendingMethod: {
    Default: 0,
    SrgbOff: 1,
  },
  EBlendingMode: {
    Normal: 0,
    Additive: 1,
    Substract: 2,
    Screen: 3,
    Multiply: 4,
    Lighten: 5,
    Darken: 6,
  },
  EFontStyle: {
    Bold: 1,
    Italic: 2,
    Underline: 4,
    Strikeout: 8,
  },
  EPropertyType: {
    Invalid: 0,
    Boolean: 1,
    Int: 2,
    Float: 3,
    Text: 4,
    Path: 5,
    List: 6,
    Color: 7,
    Button: 8,
    Font: 9,
    EditableList: 10,
    FrameRate: 11,
    Group: 12,
    ColorAlpha: 13,
    Capture: 14,
  },
  EPathType: {
    File: 0,
    FileSave: 1,
    Directory: 2,
  },
  ETextType: {
    Default: 0,
    Password: 1,
    Multiline: 2,
    TextInfo: 3,
  },
  ETextInfoType: {
    Normal: 0,
    Warning: 1,
    Error: 2,
  },
  ENumberType: {
    Scroller: 0,
    Slider: 1,
  },
  ESourceOutputFlags: {
    Video: 1,
    Audio: 2,
    Async: 4,
    AsyncVideo: 5,
    CustomDraw: 8,
    Interaction: 32,
    Composite: 64,
    DoNotDuplicate: 128,
    Deprecated: 256,
    DoNotSelfMonitor: 512,
    ForceUiRefresh: 1073741824,
  },
  ESceneDupType: {
    Refs: 0,
    Copy: 1,
    PrivateRefs: 2,
    PrivateCopy: 3,
  },
  EFaderType: {
    Cubic: 0,
    IEC: 1,
    Log: 2,
  },
  EScaleType: {
    Disable: 0,
    Point: 1,
    Bicubic: 2,
    Bilinear: 3,
    Lanczos: 4,
    Area: 5,
  },
  EFPSType: {
    Common: 0,
    Integer: 1,
    Fractional: 2,
  },
  ERangeType: {
    Default: 0,
    Partial: 1,
    Full: 2,
  },
  EVideoFormat: {
    None: 0,
    I420: 1,
    NV12: 2,
    YVYU: 3,
    YUY2: 4,
    UYVY: 5,
    RGBA: 6,
    BGRA: 7,
    BGRX: 8,
    Y800: 9,
    I444: 10,
    BGR3: 11,
    I422: 12,
    I40A: 13,
    I42A: 14,
    YUVA: 15,
    AYUV: 16,
  },
  EColorSpace: {
    Default: 0,
    CS601: 1,
    CS709: 2,
    CSSRGB: 3,
    CS2100PQ: 4,
    CS2100HLG: 5,
  },
  EOutputCode: {
    Success: 0,
    BadPath: -1,
    ConnectFailed: -2,
    InvalidStream: -3,
    Error: -4,
    Disconnected: -5,
    Unsupported: -6,
    NoSpace: -7,
    EncoderError: -8,
    OutdatedDriver: -65,
  },
  ERenderingMode: {
    OBS_MAIN_RENDERING: 0,
    OBS_STREAMING_RENDERING: 1,
    OBS_RECORDING_RENDERING: 2,
  },
  EVcamInstalledStatus: {
    NotInstalled: 0,
    LegacyInstalled: 1,
    Installed: 2,
  },
  EInteractionFlags: {
    None: 0,
    CapsKey: 1,
    ShiftKey: 2,
    ControlKey: 4,
    AltKey: 8,
    MouseLeft: 16,
    MouseMiddle: 32,
    MouseRight: 64,
    CommandKey: 128,
    Numlock_Key: 256,
    IsKeyPad: 512,
    IsLeft: 1024,
    IsRight: 2048,
  },
  EMouseButtonType: {
    Left: 0,
    Middle: 1,
    Right: 2,
  },
  ERecordingFormat: {
    MP4: 'mp4',
    FLV: 'flv',
    MOV: 'mov',
    MKV: 'mkv',
    MPEGTS: 'ts',
    HLS: 'm3u8',
  },
  ERecordingQuality: {
    Stream: 0,
    HighQuality: 1,
    HigherQuality: 2,
    Lossless: 3,
  },
  VCamOutputType: {
    Invalid: 0,
    SceneOutput: 1,
    SourceOutput: 2,
    ProgramView: 3,
    PreviewOutput: 4,
  },
};

for (const entry in runtimeEnums) {
  exports[entry] = runtimeEnums[entry];
}

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
