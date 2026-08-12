import 'package:flutter/material.dart';

class SegmentProgressBar extends StatelessWidget {
  final List segments;
  const SegmentProgressBar({super.key, required this.segments});

  @override
  Widget build(BuildContext context) => const SizedBox.shrink();
}

class ViewPointSegmentProgressBar extends StatelessWidget {
  final List segments;
  final void Function(int)? onSeek;
  const ViewPointSegmentProgressBar({super.key, required this.segments, this.onSeek});

  @override
  Widget build(BuildContext context) => const SizedBox.shrink();
}