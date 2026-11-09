import 'package:flutter/material.dart';
import '../../core/theme/theme.dart';

class DynamicPage extends StatelessWidget {
  const DynamicPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('动态')),
      body: const Center(
        child: Text('动态', style: TextStyle(color: AppTheme.textSecondary)),
      ),
    );
  }
}