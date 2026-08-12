import 'package:flutter/material.dart';

class ThumbDragDetails {
  final int seconds;
  const ThumbDragDetails(this.seconds);
}

class ProgressBar extends StatelessWidget {
  final Duration progress;
  final Duration buffered;
  final Duration total;
  final Color? progressBarColor;
  final Color? baseBarColor;
  final Color? bufferedBarColor;
  final Color? thumbColor;
  final Color? thumbGlowColor;
  final double? barHeight;
  final double? thumbRadius;
  final double? thumbGlowRadius;
  final void Function(ThumbDragDetails)? onDragStart;
  final void Function(ThumbDragDetails)? onDragUpdate;
  final void Function(int)? onSeek;

  const ProgressBar({
    super.key,
    required this.progress,
    required this.buffered,
    required this.total,
    this.progressBarColor,
    this.baseBarColor,
    this.bufferedBarColor,
    this.thumbColor,
    this.thumbGlowColor,
    this.barHeight,
    this.thumbRadius,
    this.thumbGlowRadius,
    this.onDragStart,
    this.onDragUpdate,
    this.onSeek,
  });

  @override
  Widget build(BuildContext context) => const SizedBox.shrink();
}