import 'package:flutter/material.dart';
import 'screens/home.dart';
import 'screens/menu_list.dart';
import 'screens/menu_detail.dart';
import 'screens/scan.dart';

void main() {
  runApp(const QrMenuApp());
}

class QrMenuApp extends StatelessWidget {
  const QrMenuApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'QR Menu',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF2563EB)),
        useMaterial3: true,
      ),
      initialRoute: '/',
      routes: {
        '/': (context) => const HomeScreen(),
        '/menus': (context) => const MenuListScreen(),
        '/menus/detail': (context) => const MenuDetailScreen(),
        '/scan': (context) => const ScanScreen(),
      },
    );
  }
}
