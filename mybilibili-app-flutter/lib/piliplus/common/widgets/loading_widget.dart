import 'package:flutter/material.dart';

class LoadingWidget extends StatelessWidget {
  final double? progress;
  final String? msg;

  const LoadingWidget({super.key, this.progress, this.msg});

  @override
  Widget build(BuildContext context) => const Center(child: Text('加载中...'));
}