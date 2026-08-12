import 'package:mybilibili_app_flutter/lang/constants.dart';

class DanmakuExtra {
  const DanmakuExtra();
}

class VideoDanmaku extends DanmakuExtra {
  final bool isLike;
  final int like;
  final int id;
  final int mid;
  final String text;

  const VideoDanmaku({
    this.isLike = false,
    this.like = 0,
    this.id = 0,
    this.mid = 0,
    this.text = '',
  });

  VideoDanmaku copyWith({bool? isLike, int? like}) {
    return VideoDanmaku(
      isLike: isLike ?? this.isLike,
      like: like ?? this.like,
      id: id,
      mid: mid,
      text: text,
    );
  }
}

class LiveDanmaku extends DanmakuExtra {
  final int mid;
  const LiveDanmaku({this.mid = 0});
}