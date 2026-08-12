import 'package:flutter/material.dart';

class DisabledIcon extends StatelessWidget {
  final bool? disable;
  final double? iconSize;
  final Color? color;
  final Widget? child;

  const DisabledIcon({
    super.key,
    this.disable,
    this.iconSize,
    this.color,
    this.child,
  });

  @override
  Widget build(BuildContext context) => child ?? const SizedBox.shrink();
}