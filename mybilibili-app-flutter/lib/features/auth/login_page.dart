import 'package:flutter/material.dart';
import '../../core/theme/theme.dart';

class LoginPage extends StatelessWidget {
  const LoginPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('登录')),
      body: const Center(
        child: Text('登录页', style: TextStyle(color: AppTheme.textSecondary)),
      ),
    );
  }
}