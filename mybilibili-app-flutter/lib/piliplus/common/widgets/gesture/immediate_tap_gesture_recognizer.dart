import 'package:flutter/gestures.dart';

class ImmediateTapGestureRecognizer extends GestureRecognizer {
  void Function(dynamic)? onTapDown;
  void Function(dynamic)? onTapUp;
  VoidCallback? onTapCancel;

  ImmediateTapGestureRecognizer({
    void Function(dynamic)? onTapDown,
    void Function(dynamic)? onTapUp,
    VoidCallback? onTapCancel,
  })  : onTapDown = onTapDown,
        onTapUp = onTapUp,
        onTapCancel = onTapCancel;

  @override
  void rejectGesture(int pointer) {}

  @override
  void acceptGesture(int pointer) {}

  @override
  void addPointer(PointerDownEvent event) {}
}