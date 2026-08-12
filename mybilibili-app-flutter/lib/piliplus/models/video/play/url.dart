class VideoFormat {
  final int quality;
  final String newDesc;

  VideoFormat({required this.quality, required this.newDesc});
}

class DashVideo {
  final int id;
  DashVideo({required this.id});
}

class DashData {
  final List<DashVideo> video;
  DashData({required this.video});
}

class PlayUrlModel {
  final DashData? dash;
  final List<VideoFormat>? supportFormats;
  final int? timeLength;

  PlayUrlModel({this.dash, this.supportFormats, this.timeLength});
}

import 'package:mybilibili_app_flutter/piliplus/models/common/video/video_quality.dart';

class VideoInfo {
  final VideoQuality quality;
  VideoInfo({required this.quality});
}