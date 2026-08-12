import 'package:mybilibili_app_flutter/piliplus/pages/danmaku/danmaku_model.dart';
import 'package:mybilibili_app_flutter/plugin/pl_player/controller.dart';
import 'package:flutter/material.dart';

class HeaderControl {
  static void likeDanmaku(DanmakuExtra extra, int cid) {}
  static void deleteDanmaku(int id, int cid) {}
  static void reportDanmaku(
    BuildContext context, {
    DanmakuExtra? extra,
    PlPlayerController? ctr,
  }) {}
  static void reportLiveDanmaku(
    BuildContext context, {
    int? roomId,
    String? msg,
    DanmakuExtra? extra,
    PlPlayerController? ctr,
  }) {}
}