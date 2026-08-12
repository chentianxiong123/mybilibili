import 'package:flutter/material.dart';

class Pair<T> extends StatelessWidget {
  final T first;
  final T second;
  final Widget Function(BuildContext, T)? builder;

  const Pair({super.key, required this.first, required this.second, this.builder});

  @override
  Widget build(BuildContext context) => Row(
    children: [
      if (builder != null) builder!(context, first) else Text('$first'),
      if (builder != null) builder!(context, second) else Text('$second'),
    ],
  );
}