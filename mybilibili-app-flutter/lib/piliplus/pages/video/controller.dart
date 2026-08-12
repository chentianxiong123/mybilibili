import 'package:hive_ce/hive.dart';
import 'package:mybilibili_app_flutter/piliplus/models/video/play/url.dart';
import 'package:mybilibili_app_flutter/piliplus/models/common/video/video_quality.dart';
import 'package:mybilibili_app_flutter/piliplus/models_new/video/video_detail/episode.dart';
import 'package:mybilibili_app_flutter/piliplus/utils/storage.dart';
import 'package:mybilibili_app_flutter/piliplus/utils/storage_pref.dart';
import 'package:mybilibili_app_flutter/piliplus/utils/storage_key.dart';
import 'package:mybilibili_app_flutter/plugin/pl_player/controller.dart';

class TimeBatteryMixin {
  TimeBatteryProvider get provider => TimeBatteryProvider();
  Future<void> startClock() async {}
  Future<void> stopClock() async {}
  Future<void> getBatteryLevelIfNeeded() async {}
}

class TimeBatteryProvider {
  bool muted = false;
  Future<void> startIfNeeded() async {}
}

class VideoDetailController extends TimeBatteryMixin {
  final data = PlayUrlModel();
  final vttSubtitlesIndex = ValueNotifier<int>(0);
  final showDmTrendChart = ValueNotifier<bool>(false);
  final showVP = ValueNotifier<bool>(false);
  final languages = ValueNotifier<List>([]);
  final currLang = ValueNotifier<String>('');
  final currentVideoQa = ValueNotifier<VideoQuality?>(null);
  final subtitles = <dynamic>[];
  final segmentProgressList = <dynamic>[];
  final viewPointList = <dynamic>[];
  final dmTrend = ValueNotifier<dynamic>(null);
  final showSteinEdgeInfo = ValueNotifier<bool>(false);

  bool get isUgc => true;
  bool get isPlayAll => false;
  bool get isFileSource => false;
  PlPlayerController? plPlayerController;
  int? seasonCid;
  int? get cid => 0;

  void setLanguage(String lang) {}
  void setSubtitle(int index) {}
  void updatePlayer() {}
  VideoInfo findVideoByQa(int qa) => VideoInfo(quality: VideoQuality.fromCode(qa));
}

import 'package:flutter/material.dart';