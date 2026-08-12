import 'package:flutter/material.dart';

class ViewSafeArea extends StatelessWidget {
  final bool right;
  final bool left;
  final Widget child;

  const ViewSafeArea({super.key, this.right = false, this.left = false, required this.child});

  @override
  Widget build(BuildContext context) => SafeArea(
    right: right,
    left: left,
    child: child,
  );
}