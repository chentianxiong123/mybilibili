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
    cardTheme: const CardThemeData(
      color: cardDark,
    ),
  );
}