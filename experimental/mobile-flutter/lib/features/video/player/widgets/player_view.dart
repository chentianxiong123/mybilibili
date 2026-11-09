import 'dart:async';
import 'package:flutter/material.dart';
import 'package:media_kit_video/media_kit_video.dart';
import 'package:canvas_danmaku/canvas_danmaku.dart';
import '../player_controller.dart';
import '../danmaku_model.dart';

class BilibiliPlayer extends StatefulWidget {
  final String url;
  final String? title;
  final List<DanmakuData>? danmakuList;
  final bool autoPlay;
  final bool showBackButton;
  final VoidCallback? onBack;

  const BilibiliPlayer({
    super.key,
    required this.url,
    this.title,
    this.danmakuList,
    this.autoPlay = true,
    this.showBackButton = false,
    this.onBack,
  });

  @override
  State<BilibiliPlayer> createState() => _BilibiliPlayerState();
}

class _BilibiliPlayerState extends State<BilibiliPlayer> {
  late final PlayerController _controller;
  late final VideoController _videoController;
  DanmakuController? _danmakuController;
  Timer? _controlsTimer;
  bool _danmakuEnabled = true;

  @override
  void initState() {
    super.initState();
    _controller = PlayerController();
    _videoController = VideoController(_controller.player);
    _controller.addListener(_onStateChanged);
    _controller.open(widget.url, play: widget.autoPlay);
    _startControlsTimer();
  }

  @override
  void dispose() {
    _controller.removeListener(_onStateChanged);
    _controlsTimer?.cancel();
    _controller.dispose();
    super.dispose();
  }

  void _onStateChanged(PlayerState state) {
    if (mounted) setState(() {});
    if (state.isPlaying && widget.danmakuList != null && _danmakuController != null) {
      _syncDanmaku();
    }
  }

  void _syncDanmaku() {
    final ms = _controller.state.position.inMilliseconds;
    final danmakuList = widget.danmakuList ?? [];
    for (final d in danmakuList) {
      if (d.timeMs >= ms - 200 && d.timeMs <= ms + 200) {
        _danmakuController?.addDanmaku(DanmakuContentItem(d.text, color: Color(d.color)));
      }
    }
  }

  void _startControlsTimer() {
    _controlsTimer?.cancel();
    _controlsTimer = Timer(const Duration(seconds: 4), () {
      if (mounted) _controller.hideControls();
    });
  }

  void _resetControlsTimer() {
    _controller.showControlsTemporarily();
    _startControlsTimer();
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () {
        _controller.toggleControls();
        _resetControlsTimer();
      },
      child: Stack(
        children: [
          Video(
            controller: _videoController,
            fill: Colors.black,
          ),
          if (_controller.state.isBuffering)
            const Center(child: CircularProgressIndicator(color: Color(0xFFFB7299))),
          if (_controller.state.showControls) ...[
            _buildTopBar(),
            _buildControlsOverlay(),
            _buildBottomBar(),
          ],
          if (_danmakuEnabled && _controller.state.isPlaying)
            DanmakuScreen(
              createdController: (e) => _danmakuController = e,
              option: DanmakuOption(
                fontSize: 24,
                opacity: 0.8,
                area: 0.4,
                duration: 8,
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildTopBar() {
    return Positioned(
      top: 0,
      left: 0,
      right: 0,
      child: Container(
        padding: EdgeInsets.only(top: MediaQuery.of(context).padding.top + 8, left: 16, right: 16),
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [Colors.black54, Colors.transparent],
          ),
        ),
        child: SafeArea(
          bottom: false,
          child: Row(
            children: [
              if (widget.showBackButton)
                IconButton(
                  icon: const Icon(Icons.arrow_back, color: Colors.white),
                  onPressed: widget.onBack ?? () => Navigator.of(context).pop(),
                ),
              if (widget.title != null)
                Expanded(
                  child: Text(
                    widget.title!,
                    style: const TextStyle(color: Colors.white, fontSize: 16),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              IconButton(
                icon: Icon(
                  _danmakuEnabled ? Icons.closed_caption : Icons.closed_caption_off,
                  color: Colors.white,
                ),
                onPressed: () => setState(() => _danmakuEnabled = !_danmakuEnabled),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildControlsOverlay() {
    return Positioned(
      left: 0,
      right: 0,
      top: 0,
      bottom: 0,
      child: Center(
        child: GestureDetector(
          onTap: () => _controller.togglePlayPause(),
          child: AnimatedOpacity(
            opacity: _controller.state.showControls ? 1.0 : 0.0,
            duration: const Duration(milliseconds: 200),
            child: Icon(
              _controller.state.isPlaying ? Icons.play_circle_outline : Icons.pause_circle_outline,
              color: Colors.white.withValues(alpha: 0.8),
              size: 64,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildBottomBar() {
    final state = _controller.state;
    return Positioned(
      bottom: 0,
      left: 0,
      right: 0,
      child: Container(
        padding: EdgeInsets.only(bottom: MediaQuery.of(context).padding.bottom + 8),
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.bottomCenter,
            end: Alignment.topCenter,
            colors: [Colors.black54, Colors.transparent],
          ),
        ),
        child: SafeArea(
          top: false,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              _buildProgressBar(state),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Row(
                  children: [
                    _buildTimeText(state.positionText, state.durationText),
                    const Spacer(),
                    _buildSpeedButton(),
                    _buildFullscreenButton(),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildProgressBar(PlayerState state) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: SliderTheme(
        data: SliderThemeData(
          trackHeight: 3,
          thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 6),
          overlayShape: const RoundSliderOverlayShape(overlayRadius: 12),
          activeTrackColor: const Color(0xFFFB7299),
          inactiveTrackColor: Colors.white.withValues(alpha: 0.3),
          thumbColor: const Color(0xFFFB7299),
          overlayColor: const Color(0xFFFB7299).withValues(alpha: 0.2),
        ),
        child: Slider(
          value: state.progress.clamp(0.0, 1.0),
          onChanged: (v) {
            final position = Duration(milliseconds: (state.duration.inMilliseconds * v).toInt());
            _controller.seekTo(position);
          },
          onChangeStart: (_) => _resetControlsTimer(),
        ),
      ),
    );
  }

  Widget _buildTimeText(String position, String duration) {
    return Text(
      '$position / $duration',
      style: const TextStyle(color: Colors.white, fontSize: 12),
    );
  }

  Widget _buildSpeedButton() {
    return GestureDetector(
      onTap: () {
        final speeds = [0.5, 0.75, 1.0, 1.25, 1.5, 2.0];
        final currentIndex = speeds.indexOf(_controller.state.speed);
        final nextSpeed = speeds[(currentIndex + 1) % speeds.length];
        _controller.setSpeed(nextSpeed);
        _resetControlsTimer();
      },
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
        decoration: BoxDecoration(
          color: Colors.white.withValues(alpha: 0.2),
          borderRadius: BorderRadius.circular(4),
        ),
        child: Text(
          '${_controller.state.speed}x',
          style: const TextStyle(color: Colors.white, fontSize: 12),
        ),
      ),
    );
  }

  Widget _buildFullscreenButton() {
    return IconButton(
      icon: Icon(
        _controller.state.isFullscreen ? Icons.fullscreen_exit : Icons.fullscreen,
        color: Colors.white,
        size: 20,
      ),
      onPressed: () => _controller.toggleFullscreen(),
      padding: EdgeInsets.zero,
      constraints: const BoxConstraints(minWidth: 36, minHeight: 36),
    );
  }
}