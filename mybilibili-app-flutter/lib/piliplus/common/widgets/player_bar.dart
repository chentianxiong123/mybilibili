import 'package:flutter/material.dart';

class PlayerBar extends StatelessWidget {
  final List<Widget> children;

  const PlayerBar({super.key, required this.children});

  @override
  Widget build(BuildContext context) => Row(
    mainAxisAlignment: MainAxisAlignment.spaceBetween,
    children: children,
  );
}