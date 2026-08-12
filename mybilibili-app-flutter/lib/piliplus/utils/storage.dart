import 'package:hive_ce/hive.dart';

class GStorage {
  static Box setting = Box<String>();
  static Box video = Box<String>();
}

class SettingBoxKey {
  static const superResolutionType = 'superResolutionType';
  static const desktopVolume = 'desktopVolume';
  static const enableBackgroundPlay = 'enableBackgroundPlay';
  static const subtitleFontScale = 'subtitleFontScale';
  static const subtitleFontScaleFS = 'subtitleFontScaleFS';
  static const subtitlePaddingH = 'subtitlePaddingH';
  static const subtitlePaddingB = 'subtitlePaddingB';
  static const subtitleBgOpacity = 'subtitleBgOpacity';
  static const subtitleStrokeWidth = 'subtitleStrokeWidth';
  static const subtitleFontWeight = 'subtitleFontWeight';
  static const continuePlayInBackground = 'continuePlayInBackground';
  static const defaultVideoQa = 'defaultVideoQa';
  static const defaultVideoQaCellular = 'defaultVideoQaCellular';
  static const danmakuBlockType = 'danmakuBlockType';
  static const danmakuShowArea = 'danmakuShowArea';
  static const danmakuFontScale = 'danmakuFontScale';
  static const danmakuFontScaleFS = 'danmakuFontScaleFS';
  static const danmakuDuration = 'danmakuDuration';
  static const danmakuStaticDuration = 'danmakuStaticDuration';
  static const danmakuStrokeWidth = 'danmakuStrokeWidth';
  static const danmakuFontWeight = 'danmakuFontWeight';
  static const danmakuLineHeight = 'danmakuLineHeight';
  static const danmakuMassiveMode = 'danmakuMassiveMode';
  static const danmakuStatic2Scroll = 'danmakuStatic2Scroll';
  static const danmakuFixedV = 'danmakuFixedV';
  static const danmakuWeight = 'danmakuWeight';
  static const danmakuOpacity = 'danmakuOpacity';
}

class VideoBoxKey {
  static const cacheVideoFit = 'cacheVideoFit';
  static const playRepeat = 'playRepeat';
}

import 'package:hive_ce/hive.dart';