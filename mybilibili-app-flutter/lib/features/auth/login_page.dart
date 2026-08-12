import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/auth_api.dart';
import '../../core/theme/theme.dart';

class LoginPage extends ConsumerStatefulWidget {
  const LoginPage({super.key});
  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage>
    with SingleTickerProviderStateMixin {
  late final TabController _tabController;
  final _usernameController = TextEditingController();
  final _passwordController = TextEditingController();
  final _nicknameController = TextEditingController();
  final _emailController = TextEditingController();
  bool _loading = false;
  bool _obscurePassword = true;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    _usernameController.dispose();
    _passwordController.dispose();
    _nicknameController.dispose();
    _emailController.dispose();
    super.dispose();
  }

  Future<void> _login() async {
    if (_usernameController.text.trim().isEmpty || _passwordController.text.isEmpty) {
      _showMessage('请输入用户名和密码');
      return;
    }
    setState(() => _loading = true);
    try {
      await ref.read(authApiProvider).login(
        _usernameController.text.trim(),
        _passwordController.text,
      );
      if (mounted) {
        _showMessage('登录成功', isError: false);
        Navigator.of(context).pop();
      }
    } catch (e) {
      if (mounted) _showMessage('登录失败: $e');
    }
    if (mounted) setState(() => _loading = false);
  }

  Future<void> _register() async {
    if (_usernameController.text.trim().isEmpty || _passwordController.text.isEmpty) {
      _showMessage('请输入用户名和密码');
      return;
    }
    setState(() => _loading = true);
    try {
      await ref.read(authApiProvider).register(
        username: _usernameController.text.trim(),
        password: _passwordController.text,
        nickname: _nicknameController.text.trim().isNotEmpty ? _nicknameController.text.trim() : null,
        email: _emailController.text.trim().isNotEmpty ? _emailController.text.trim() : null,
      );
      if (mounted) {
        _showMessage('注册成功', isError: false);
        Navigator.of(context).pop();
      }
    } catch (e) {
      if (mounted) _showMessage('注册失败: $e');
    }
    if (mounted) setState(() => _loading = false);
  }

  void _showMessage(String msg, {bool isError = true}) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: isError ? Colors.red : AppTheme.primaryPink,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('账号'),
        bottom: TabBar(
          controller: _tabController,
          labelColor: AppTheme.primaryPink,
          unselectedLabelColor: Colors.grey,
          indicatorColor: AppTheme.primaryPink,
          tabs: const [Tab(text: '登录'), Tab(text: '注册')],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [_buildLoginForm(), _buildRegisterForm()],
      ),
    );
  }

  Widget _buildInput(TextEditingController controller, String hint, {bool obscure = false, bool enabled = true}) {
    return Container(
      height: 44,
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: const Color(0xFF2A2A2A),
        borderRadius: BorderRadius.circular(22),
      ),
      child: TextField(
        controller: controller,
        obscureText: obscure,
        enabled: enabled,
        style: const TextStyle(fontSize: 14),
        decoration: InputDecoration(
          hintText: hint,
          hintStyle: const TextStyle(fontSize: 14, color: Colors.grey),
          border: InputBorder.none,
          contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
        ),
      ),
    );
  }

  Widget _buildLoginForm() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        children: [
          const SizedBox(height: 16),
          _buildInput(_usernameController, '用户名'),
          _buildInput(_passwordController, '密码', obscure: true),
          SizedBox(
            width: double.infinity,
            height: 44,
            child: ElevatedButton(
              onPressed: _loading ? null : _login,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppTheme.primaryPink,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(22)),
              ),
              child: _loading
                  ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                  : const Text('登录', style: TextStyle(fontSize: 15, color: Colors.white)),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRegisterForm() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        children: [
          const SizedBox(height: 16),
          _buildInput(_usernameController, '用户名'),
          _buildInput(_passwordController, '密码', obscure: true),
          _buildInput(_nicknameController, '昵称（可选）'),
          _buildInput(_emailController, '邮箱（可选）'),
          SizedBox(
            width: double.infinity,
            height: 44,
            child: ElevatedButton(
              onPressed: _loading ? null : _register,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppTheme.primaryPink,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(22)),
              ),
              child: _loading
                  ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                  : const Text('注册', style: TextStyle(fontSize: 15, color: Colors.white)),
            ),
          ),
        ],
      ),
    );
  }
}