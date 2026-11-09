import 'package:media_kit/media_kit.dart';

class PlayerState {
  final bool isPlaying;
  final bool isPaused;
  final bool isComplete;
  final Duration position;
  final Duration duration;
  final double volume;
  final double speed;
  final bool isFullscreen;
  final bool isBuffering;
  final bool showControls;

  const PlayerState({
    this.isPlaying = false,
    this.isPaused = false,
    this.isComplete = false,
    this.position = Duration.zero,
    this.duration = Duration.zero,
    this.volume = 1.0,
    this.speed = 1.0,
    this.isFullscreen = false,
    this.isBuffering = false,
    this.showControls = true,
  });

  PlayerState copyWith({
    bool? isPlaying,
    bool? isPaused,
    bool? isComplete,
    Duration? position,
    Duration? duration,
    double? volume,
    double? speed,
    bool? isFullscreen,
    bool? isBuffering,
    bool? showControls,
  }) {
    return PlayerState(
      isPlaying: isPlaying ?? this.isPlaying,
      isPaused: isPaused ?? this.isPaused,
      isComplete: isComplete ?? this.isComplete,
      position: position ?? this.position,
      duration: duration ?? this.duration,
      volume: volume ?? this.volume,
      speed: speed ?? this.speed,
      isFullscreen: isFullscreen ?? this.isFullscreen,
      isBuffering: isBuffering ?? this.isBuffering,
      showControls: showControls ?? this.showControls,
    );
  }

  double get progress => duration.inMilliseconds > 0
      ? position.inMilliseconds / duration.inMilliseconds
      : 0.0;

  String get positionText {
    final h = position.inHours;
    final m = position.inMinutes.remainder(60);
    final s = position.inSeconds.remainder(60);
    if (h > 0) {
      return '$h:${m.toString().padLeft(2, '0')}:${s.toString().padLeft(2, '0')}';
    }
    return '$m:${s.toString().padLeft(2, '0')}';
  }

  String get durationText {
    final h = duration.inHours;
    final m = duration.inMinutes.remainder(60);
    final s = duration.inSeconds.remainder(60);
    if (h > 0) {
      return '$h:${m.toString().padLeft(2, '0')}:${s.toString().padLeft(2, '0')}';
    }
    return '$m:${s.toString().padLeft(2, '0')}';
  }
}

class PlayerController {
  final Player player = Player();
  PlayerState state = const PlayerState();
  final List<void Function(PlayerState)> listeners = [];

  PlayerController();

  void addListener(void Function(PlayerState) listener) {
    listeners.add(listener);
  }

  void removeListener(void Function(PlayerState) listener) {
    listeners.remove(listener);
  }

  void _notify() {
    for (final l in listeners) {
      l(state);
    }
  }

  void _updateState(PlayerState newState) {
    state = newState;
    _notify();
  }

  Future<void> open(String url, {bool play = true}) async {
    await player.open(Media(url), play: play);
    player.stream.position.listen((p) {
      _updateState(state.copyWith(position: p));
    });
    player.stream.duration.listen((d) {
      _updateState(state.copyWith(duration: d));
    });
    player.stream.playing.listen((p) {
      _updateState(state.copyWith(isPlaying: p, isPaused: !p));
    });
    player.stream.completed.listen((_) {
      _updateState(state.copyWith(isComplete: true, isPlaying: false));
    });
    player.stream.buffering.listen((b) {
      _updateState(state.copyWith(isBuffering: b));
    });
  }

  Future<void> play() async {
    await player.play();
    _updateState(state.copyWith(isPlaying: true, isPaused: false));
  }

  Future<void> pause() async {
    await player.pause();
    _updateState(state.copyWith(isPlaying: false, isPaused: true));
  }

  Future<void> togglePlayPause() async {
    if (state.isPlaying) {
      await pause();
    } else {
      await play();
    }
  }

  Future<void> seekTo(Duration position) async {
    await player.seek(position);
    _updateState(state.copyWith(position: position));
  }

  Future<void> setVolume(double volume) async {
    await player.setVolume(volume);
    _updateState(state.copyWith(volume: volume));
  }

  Future<void> setSpeed(double speed) async {
    await player.setRate(speed);
    _updateState(state.copyWith(speed: speed));
  }

  Future<void> toggleFullscreen() async {
    _updateState(state.copyWith(isFullscreen: !state.isFullscreen));
  }

  void showControlsTemporarily() {
    _updateState(state.copyWith(showControls: true));
  }

  void hideControls() {
    _updateState(state.copyWith(showControls: false));
  }

  void toggleControls() {
    _updateState(state.copyWith(showControls: !state.showControls));
  }

  Future<void> dispose() async {
    await player.dispose();
  }
}