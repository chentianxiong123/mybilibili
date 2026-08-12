import 'package:flutter/material.dart';
import '../../core/theme/theme.dart';

class HotPage extends StatelessWidget {
  const HotPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('热门')),
      body: const Center(
        child: Text('热门', style: TextStyle(color: AppTheme.textSecondary)),
      ),
    );
  }
}