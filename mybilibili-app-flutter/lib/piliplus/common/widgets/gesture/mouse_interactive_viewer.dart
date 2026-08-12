import 'package:flutter/material.dart';

class MouseInteractiveViewer extends StatelessWidget {
  final bool scaleEnabled;
  final bool pointerSignalFallback;
  final void Function(dynamic)? onPointerPanZoomUpdate;
  final void Function(dynamic)? onPointerPanZoomEnd;
  final void Function(dynamic)? onPointerDown;
  final void Function(dynamic)? onPanStart;
  final void Function(dynamic)? onPanUpdate;
  final void Function(dynamic)? onPanEnd;
  final void Function(dynamic)? onScaleUpdate;
  final GestureRecognizer? scaleGestureRecognizer;
  final bool panEnabled;
  final double minScale;
  final double maxScale;
  final EdgeInsets boundaryMargin;
  final PanAxis panAxis;
  final TransformationController? transformationController;
  final Key? childKey;
  final Widget child;

  const MouseInteractiveViewer({
    super.key,
    this.scaleEnabled = true,
    this.pointerSignalFallback = false,
    this.onPointerPanZoomUpdate,
    this.onPointerPanZoomEnd,
    this.onPointerDown,
    this.onPanStart,
    this.onPanUpdate,
    this.onPanEnd,
    this.onScaleUpdate,
    this.scaleGestureRecognizer,
    this.panEnabled = true,
    this.minScale = 1.0,
    this.maxScale = 5.0,
    this.boundaryMargin = EdgeInsets.zero,
    this.panAxis = PanAxis.free,
    this.transformationController,
    this.childKey,
    required this.child,
  });

  @override
  Widget build(BuildContext context) => child;
}