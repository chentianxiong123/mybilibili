# S1: Flutter App Skeleton Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create the Flutter project scaffold with pl_player integrated, core infrastructure set up, and a working app shell with bottom navigation.

**Architecture:** Flutter 3.44.9 + Dart 3.12.2. Extract pl_player from PiliPlus as a local plugin. Use Riverpod for state management, GoRouter for navigation, Dio for HTTP. Follow feature-first directory structure.

**Tech Stack:** Flutter 3.44.9, media_kit 1.1.11, canvas_danmaku, dio, riverpod, go_router, flutter_secure_storage

---

### Task 1: Create Flutter Project

**Files:**
- Create: `mybilibili-app-flutter/` (via `flutter create`)

- [ ] **Step 1: Create the Flutter project**

```bash
cd /tmp/mybilibili
flutter create --org com.mybilibili --project-name mybilibili_app_flutter mybilibili-app-flutter
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /tmp/mybilibili/mybilibili-app-flutter
flutter analyze
```
Expected: "No issues found"

- [ ] **Step 3: Commit**

```bash
cd /tmp/mybilibili
git add mybilibili-app-flutter/
git commit -m "feat: scaffold Flutter app"
```

---

### Task 2: Configure pubspec.yaml with Dependencies

**Files:**
- Modify: `mybilibili-app-flutter/pubspec.yaml`

- [ ] **Step 1: Replace pubspec.yaml content**

```yaml
name: mybilibili_app_flutter
description: "mybilibili Flutter client - Bilibili-style video platform"
publish_to: 'none'
version: 1.0.0+1

environment:
  sdk: ^3.12.0

dependencies:
  flutter:
    sdk: flutter
  cupertino_icons: ^1.0.8

  # State management
  flutter_riverpod: ^2.6.1
  riverpod_annotation: ^2.6.1

  # Navigation
  go_router: ^14.8.1

  # HTTP
  dio: ^5.7.0

  # Video player
  media_kit: 1.1.11
  media_kit_video: 1.2.5
  media_kit_libs_video: 1.0.5

  # Danmaku
  canvas_danmaku: ^0.3.1

  # Storage
  flutter_secure_storage: ^9.2.4
  hive: ^2.2.3
  shared_preferences: ^2.3.4

  # DLNA
  dlna_dart: ^0.1.0

  # UI
  flutter_svg: ^2.0.16
  cached_network_image: ^3.4.1
  shimmer: ^3.0.0
  intl: ^0.19.0

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^5.0.0
  build_runner: ^2.4.14
  riverpod_generator: ^2.6.3

flutter:
  uses-material-design: true
```

- [ ] **Step 2: Run flutter pub get**

```bash
cd /tmp/mybilibili/mybilibili-app-flutter
flutter pub get
```
Expected: "Process finished with exit code 0"

- [ ] **Step 3: Commit**

```bash
cd /tmp/mybilibili
git add mybilibili-app-flutter/pubspec.yaml mybilibili-app-flutter/pubspec.lock
git commit -m "feat: add dependencies (media_kit, riverpod, dio, canvas_danmaku)"
```

---

### Task 3: Extract pl_player from PiliPlus

**Files:**
- Create: `mybilibili-app-flutter/lib/plugin/pl_player/` (all files from PiliPlus)

- [ ] **Step 1: Copy pl_player directory**

```bash
cd /tmp/mybilibili/mybilibili-app-flutter
cp -r /tmp/pp-ref/lib/plugin/pl_player lib/plugin/
```

- [ ] **Step 2: Remove PiliPlus-specific imports (replace with local equivalents)**

The pl_player files import from `package:PiliPlus/...` which doesn't exist in our project. We need to replace these with local stubs or remove them. For now, we'll create stub files for the most critical dependencies.

```bash
# Create stubs directory for PiliPlus references
mkdir -p lib/plugin/pl_player/stubs
```

- [ ] **Step 3: Create stub files for PiliPlus-specific imports**

Create `lib/plugin/pl_player/stubs/common.dart`:
```dart
// Stub file replacing package:PiliPlus/common/* imports
// Will be replaced with our own implementations
```

- [ ] **Step 4: Verify pl_player compiles (or at least analyze errors)**

```bash
cd /tmp/mybilibili/mybilibili-app-flutter
flutter analyze lib/plugin/pl_player/
```
Expected: List of errors (expected - we need to fix imports)

- [ ] **Step 5: Commit**

```bash
cd /tmp/mybilibili
git add mybilibili-app-flutter/lib/plugin/
git commit -m "feat: extract pl_player from PiliPlus"
```

---

### Task 4: Create Core Infrastructure

**Files:**
- Create: `mybilibili-app-flutter/lib/core/api/client.dart`
- Create: `mybilibili-app-flutter/lib/core/api/auth_api.dart`
- Create: `mybilibili-app-flutter/lib/core/api/manuscript_api.dart`
- Create: `mybilibili-app-flutter/lib/core/theme/theme.dart`
- Create: `mybilibili-app-flutter/lib/core/utils/token_storage.dart`

- [ ] **Step 1: Create API client**

`lib/core/api/client.dart`:
```dart
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../utils/token_storage.dart';

final dioProvider = Provider<Dio>((ref) {
  final dio = Dio(BaseOptions(
    baseUrl: 'http://localhost:8080/api/v1',
    connectTimeout: const Duration(seconds: 10),
    receiveTimeout: const Duration(seconds: 10),
    headers: {'Content-Type': 'application/json'},
  ));

  dio.interceptors.add(InterceptorsWrapper(
    onRequest: (options, handler) async {
      final token = await TokenStorage.getAccessToken();
      if (token != null) {
        options.headers['Authorization'] = 'Bearer $token';
      }
      handler.next(options);
    },
    onError: (error, handler) async {
      if (error.response?.statusCode == 401) {
        // Token refresh logic
        final refreshToken = await TokenStorage.getRefreshToken();
        if (refreshToken != null) {
          try {
            final response = await Dio().post(
              '${dio.options.baseUrl}/user/token/refresh',
              data: {'refreshToken': refreshToken},
            );
            final newToken = response.data['data']['accessToken'];
            await TokenStorage.saveAccessToken(newToken);
            error.requestOptions.headers['Authorization'] = 'Bearer $newToken';
            final retryResponse = await dio.fetch(error.requestOptions);
            handler.resolve(retryResponse);
            return;
          } catch (_) {}
        }
        // If refresh fails, redirect to login
        await TokenStorage.clearTokens();
      }
      handler.next(error);
    },
  ));

  return dio;
});
```

- [ ] **Step 2: Create token storage utility**

`lib/core/utils/token_storage.dart`:
```dart
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class TokenStorage {
  static const _storage = FlutterSecureStorage();
  static const _accessTokenKey = 'access_token';
  static const _refreshTokenKey = 'refresh_token';
  static const _userIdKey = 'user_id';

  static Future<void> saveAccessToken(String token) =>
      _storage.write(key: _accessTokenKey, value: token);

  static Future<String?> getAccessToken() =>
      _storage.read(key: _accessTokenKey);

  static Future<void> saveRefreshToken(String token) =>
      _storage.write(key: _refreshTokenKey, value: token);

  static Future<String?> getRefreshToken() =>
      _storage.read(key: _refreshTokenKey);

  static Future<void> saveUserId(int id) =>
      _storage.write(key: _userIdKey, value: id.toString());

  static Future<int?> getUserId() async {
    final id = await _storage.read(key: _userIdKey);
    return id != null ? int.tryParse(id) : null;
  }

  static Future<void> clearTokens() async {
    await _storage.delete(key: _accessTokenKey);
    await _storage.delete(key: _refreshTokenKey);
    await _storage.delete(key: _userIdKey);
  }
}
```

- [ ] **Step 3: Create theme**

`lib/core/theme/theme.dart`:
```dart
import 'package:flutter/material.dart';

class AppTheme {
  static const Color primaryPink = Color(0xFFFB7299);
  static const Color primaryPinkDark = Color(0xFFE85C8A);
  static const Color backgroundDark = Color(0xFF181818);
  static const Color surfaceDark = Color(0xFF222222);
  static const Color cardDark = Color(0xFF2A2A2A);
  static const Color textPrimary = Color(0xFFFFFFFF);
  static const Color textSecondary = Color(0xFF99A2AA);

  static ThemeData get darkTheme => ThemeData(
    brightness: Brightness.dark,
    primaryColor: primaryPink,
    scaffoldBackgroundColor: backgroundDark,
    colorScheme: const ColorScheme.dark(
      primary: primaryPink,
      secondary: primaryPinkDark,
      surface: surfaceDark,
    ),
    appBarTheme: const AppBarTheme(
      backgroundColor: backgroundDark,
      elevation: 0,
      centerTitle: true,
    ),
    bottomNavigationBarTheme: const BottomNavigationBarThemeData(
      backgroundColor: surfaceDark,
      selectedItemColor: primaryPink,
      unselectedItemColor: textSecondary,
    ),
  );
}
```

- [ ] **Step 4: Create auth API**

`lib/core/api/auth_api.dart`:
```dart
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';

final authApiProvider = Provider<AuthApi>((ref) {
  return AuthApi(ref.read(dioProvider));
});

class AuthApi {
  final Dio _dio;
  AuthApi(this._dio);

  Future<Map<String, dynamic>> login(String username, String password) async {
    final response = await _dio.post('/user/login', data: {
      'username': username,
      'password': password,
    });
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> register(String username, String password, String email) async {
    final response = await _dio.post('/user/register', data: {
      'username': username,
      'password': password,
      'email': email,
    });
    return response.data as Map<String, dynamic>;
  }
}
```

- [ ] **Step 5: Create manuscript API**

`lib/core/api/manuscript_api.dart`:
```dart
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';

final manuscriptApiProvider = Provider<ManuscriptApi>((ref) {
  return ManuscriptApi(ref.read(dioProvider));
});

class ManuscriptApi {
  final Dio _dio;
  ManuscriptApi(this._dio);

  Future<Map<String, dynamic>> getRecommended({int page = 1, int pageSize = 20}) async {
    final response = await _dio.get('/manuscript/recommended', queryParameters: {
      'page': page,
      'pageSize': pageSize,
    });
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getManuscript(int id) async {
    final response = await _dio.get('/manuscript/$id');
    return response.data as Map<String, dynamic>;
  }
}
```

- [ ] **Step 6: Commit**

```bash
cd /tmp/mybilibili
git add mybilibili-app-flutter/lib/core/
git commit -m "feat: add core infrastructure (Dio, theme, token storage, APIs)"
```

---

### Task 5: Create App Shell with Navigation

**Files:**
- Create: `mybilibili-app-flutter/lib/app.dart`
- Modify: `mybilibili-app-flutter/lib/main.dart`
- Create: `mybilibili-app-flutter/lib/core/router/router.dart`
- Create: `mybilibili-app-flutter/lib/features/home/home_page.dart`
- Create: `mybilibili-app-flutter/lib/features/home/dynamic_page.dart`
- Create: `mybilibili-app-flutter/lib/features/home/hot_page.dart`
- Create: `mybilibili-app-flutter/lib/features/auth/login_page.dart`
- Create: `mybilibili-app-flutter/lib/features/profile/profile_page.dart`

- [ ] **Step 1: Create router**

`lib/core/router/router.dart`:
```dart
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../features/home/home_page.dart';
import '../../features/home/dynamic_page.dart';
import '../../features/home/hot_page.dart';
import '../../features/auth/login_page.dart';
import '../../features/profile/profile_page.dart';

final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/home',
    routes: [
      ShellRoute(
        builder: (context, state, child) => AppShell(child: child),
        routes: [
          GoRoute(
            path: '/home',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: HomePage(),
            ),
          ),
          GoRoute(
            path: '/hot',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: HotPage(),
            ),
          ),
          GoRoute(
            path: '/dynamic',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: DynamicPage(),
            ),
          ),
          GoRoute(
            path: '/profile',
            pageBuilder: (context, state) => const NoTransitionPage(
              child: ProfilePage(),
            ),
          ),
        ],
      ),
      GoRoute(path: '/login', builder: (context, state) => const LoginPage()),
    ],
  );
});

class AppShell extends StatelessWidget {
  final Widget child;
  const AppShell({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: child,
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _calculateIndex(context),
        onTap: (index) => _onTap(context, index),
        items: const [
          BottomNavigationBarItem(icon: Icon(Icons.home), label: '首页'),
          BottomNavigationBarItem(icon: Icon(Icons.local_fire_department), label: '热门'),
          BottomNavigationBarItem(icon: Icon(Icons.dynamic_feed), label: '动态'),
          BottomNavigationBarItem(icon: Icon(Icons.person), label: '我的'),
        ],
      ),
    );
  }

  int _calculateIndex(BuildContext context) {
    final location = GoRouterState.of(context).uri.toString();
    if (location.startsWith('/home')) return 0;
    if (location.startsWith('/hot')) return 1;
    if (location.startsWith('/dynamic')) return 2;
    if (location.startsWith('/profile')) return 3;
    return 0;
  }

  void _onTap(BuildContext context, int index) {
    switch (index) {
      case 0: context.go('/home');
      case 1: context.go('/hot');
      case 2: context.go('/dynamic');
      case 3: context.go('/profile');
    }
  }
}
```

- [ ] **Step 2: Create main.dart with ProviderScope**

`lib/main.dart`:
```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'app.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const ProviderScope(child: MyBiliApp()));
}
```

- [ ] **Step 3: Create app.dart**

`lib/app.dart`:
```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'core/router/router.dart';
import 'core/theme/theme.dart';

class MyBiliApp extends ConsumerWidget {
  const MyBiliApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);
    return MaterialApp.router(
      title: 'mybilibili',
      theme: AppTheme.darkTheme,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
```

- [ ] **Step 4: Create placeholder pages**

`lib/features/home/home_page.dart`:
```dart
import 'package:flutter/material.dart';
import '../../core/theme/theme.dart';

class HomePage extends StatelessWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('mybilibili')),
      body: const Center(
        child: Text('首页 - 推荐', style: TextStyle(color: AppTheme.textSecondary)),
      ),
    );
  }
}
```

`lib/features/home/dynamic_page.dart`:
```dart
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
```

`lib/features/home/hot_page.dart`:
```dart
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
```

`lib/features/auth/login_page.dart`:
```dart
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
```

`lib/features/profile/profile_page.dart`:
```dart
import 'package:flutter/material.dart';
import '../../core/theme/theme.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('我的')),
      body: const Center(
        child: Text('个人中心', style: TextStyle(color: AppTheme.textSecondary)),
      ),
    );
  }
}
```

- [ ] **Step 5: Fix import in router.dart (add Provider import)**

Add to `lib/core/router/router.dart`:
```dart
import 'package:flutter_riverpod/flutter_riverpod.dart';
```

- [ ] **Step 6: Run flutter analyze to verify**

```bash
cd /tmp/mybilibili/mybilibili-app-flutter
flutter analyze
```
Expected: "No issues found" (or minimal warnings)

- [ ] **Step 7: Commit**

```bash
cd /tmp/mybilibili
git add mybilibili-app-flutter/lib/app.dart mybilibili-app-flutter/lib/main.dart mybilibili-app-flutter/lib/core/router/ mybilibili-app-flutter/lib/features/
git commit -m "feat: add app shell with bottom navigation"
```

---

### Task 6: Fix pl_player Imports and Make It Compile

**Files:**
- Modify: `mybilibili-app-flutter/lib/plugin/pl_player/` (all files)

- [ ] **Step 1: Analyze pl_player to get full error list**

```bash
cd /tmp/mybilibili/mybilibili-app-flutter
flutter analyze lib/plugin/pl_player/ 2>&1 | tee /tmp/pl_player_errors.txt
```

- [ ] **Step 2: Fix each import error systematically**

The pl_player files import from many PiliPlus-specific packages:
- `package:PiliPlus/common/...`
- `package:PiliPlus/models/...`
- `package:PiliPlus/pages/...`
- `package:PiliPlus/utils/...`
- `package:get/get.dart`
- `package:flutter_smart_dialog/...`
- `package:flutter_volume_controller/...`
- `package:screen_brightness_platform_interface/...`
- `package:PiliPlus/common/widgets/...`

Strategy: Replace each import with local stubs or remove the dependency. For now, we'll create minimal stubs that allow compilation.

- [ ] **Step 3: Create stub package for PiliPlus references**

`lib/plugin/pl_player/stubs/common.dart`:
```dart
// Stub file for PiliPlus common imports
// TODO: Replace with actual implementations

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class Utils {
  static bool get isDesktop => false;
  static bool get isMobile => true;
  static void copyText(String text) {
    Clipboard.setData(ClipboardData(text: text));
  }
}

class Pref {
  static bool get enableTapDm => true;
}

class StringExt {
  static String secToTime(int seconds) {
    final d = Duration(seconds: seconds);
    return '${d.inMinutes.toString().padLeft(2, '0')}:${(d.inSeconds % 60).toString().padLeft(2, '0')}';
  }
}
```

- [ ] **Step 4: Iteratively fix all import errors**

For each file in pl_player, replace `package:PiliPlus/...` imports with `package:mybilibili_app_flutter/plugin/pl_player/stubs/...` imports.

- [ ] **Step 5: Verify pl_player compiles**

```bash
cd /tmp/mybilibili/mybilibili-app-flutter
flutter analyze lib/plugin/pl_player/controller.dart
```
Expected: "No issues found"

- [ ] **Step 6: Full project analyze**

```bash
cd /tmp/mybilibili/mybilibili-app-flutter
flutter analyze
```
Expected: "No issues found"

- [ ] **Step 7: Commit**

```bash
cd /tmp/mybilibili
git add mybilibili-app-flutter/lib/plugin/
git commit -m "fix: adapt pl_player imports for standalone project"
```