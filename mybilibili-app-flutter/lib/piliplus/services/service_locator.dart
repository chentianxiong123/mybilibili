class VideoPlayerServiceHandler {
  Future<void> onStatusChange(dynamic status) async {}
  Future<void> onPositionChange(dynamic position) async {}
  bool? enableBackgroundPlay;
  Future<void> clear() async {}
}

class AudioSessionHandler {
  Future<void> setActive(bool active) async {}
}

final videoPlayerServiceHandler = VideoPlayerServiceHandler();
final audioSessionHandler = AudioSessionHandler();