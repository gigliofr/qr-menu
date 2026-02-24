import 'package:flutter/material.dart';
import '../models/menu.dart';

class MenuListScreen extends StatelessWidget {
  const MenuListScreen({super.key});

  List<Menu> get menus => const [
        Menu(
          id: '1',
          name: 'Lunch Menu',
          description: 'Seasonal favorites and specials',
          categories: [],
        ),
        Menu(
          id: '2',
          name: 'Dinner Menu',
          description: 'Evening dishes and pairings',
          categories: [],
        ),
        Menu(
          id: '3',
          name: 'Desserts',
          description: 'Sweet finishes',
          categories: [],
        ),
      ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Menus'),
      ),
      body: ListView.separated(
        padding: const EdgeInsets.all(16),
        itemBuilder: (context, index) {
          final menu = menus[index];
          return ListTile(
            tileColor: Colors.white,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
              side: const BorderSide(color: Color(0xFFE2E8F0)),
            ),
            title: Text(menu.name, style: Theme.of(context).textTheme.titleMedium),
            subtitle: Text(menu.description),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => Navigator.pushNamed(context, '/menus/detail', arguments: menu),
          );
        },
        separatorBuilder: (_, __) => const SizedBox(height: 12),
        itemCount: menus.length,
      ),
    );
  }
}
