import 'package:flutter/material.dart';

class PopupMenuText<T> extends StatelessWidget {
  final String title;
  final T value;
  final void Function(T)? onSelected;
  final List<PopupMenuEntry<T>> Function()? itemBuilder;
  final String Function(T)? getSelectTitle;
  final Widget? child;

  const PopupMenuText({
    super.key,
    required this.title,
    required this.value,
    this.onSelected,
    this.itemBuilder,
    this.getSelectTitle,
    this.child,
  });

  @override
  Widget build(BuildContext context) => Text(title);
}