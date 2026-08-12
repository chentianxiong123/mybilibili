import 'package:mybilibili_app_flutter/piliplus/models_new/video/video_detail/episode.dart';
import 'package:mybilibili_app_flutter/piliplus/models/common/video/video_quality.dart';
import 'package:mybilibili_app_flutter/piliplus/models/video/play/url.dart';
import 'package:flutter/material.dart';

class VideoDetail {
  UgcSeason? ugcSeason;
  List<BaseEpisodeItem>? pages;
}

class CommonIntroController {
  final videoDetail = ValueNotifier<VideoDetail?>(null);
  Future<void> prevPlay() async {}
  Future<void> nextPlay() async {}
}

import 'package:flutter_riverpod/flutter_riverpod.dart';