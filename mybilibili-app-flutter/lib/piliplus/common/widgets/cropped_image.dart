import 'package:flutter/material.dart';

class CroppedImage extends StatelessWidget {
  final double? size;
  final ImageProvider? image;
  final Rect? srcRect;
  final Rect? dstRect;
  final RRect? rrect;
  final Paint? imgPaint;
  final Paint? borderPaint;

  const CroppedImage({
    super.key,
    this.size,
    this.image,
    this.srcRect,
    this.dstRect,
    this.rrect,
    this.imgPaint,
    this.borderPaint,
  });

  @override
  Widget build(BuildContext context) => const SizedBox.shrink();
}