import 'package:flutter/gestures.dart';

class PlayerScaleGestureRecognizer extends ScaleGestureRecognizer {
  bool isPosAllowed = false;

  PlayerScaleGestureRecognizer({
    Object? debugOwner,
    DragStartBehavior dragStartBehavior = DragStartBehavior.start,
    Set<PointerDeviceKind>? allowedButtonsFilter,
    double? trackpadScrollToScaleFactor,
    bool? trackpadScrollCausesScale,
  }) : super(debugOwner: debugOwner);
}