import 'package:flutter/material.dart';
import '../models/menu.dart';

class MenuDetailScreen extends StatelessWidget {
  const MenuDetailScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final menu = ModalRoute.of(context)?.settings.arguments as Menu?;

    return Scaffold(
      appBar: AppBar(
        title: Text(menu?.name ?? 'Menu Detail'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(menu?.description ?? 'Menu description',
                style: Theme.of(context).textTheme.bodyLarge),
            const SizedBox(height: 24),
            Text('Categories', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 12),
            Expanded(
              child: ListView.separated(
                itemCount: menu?.categories.length ?? 0,
                itemBuilder: (context, index) {
                  final category = menu!.categories[index];
                  return ListTile(
                    title: Text(category.name),
                    subtitle: Text('${category.items.length} items'),
                  );
                },
                separatorBuilder: (_, __) => const Divider(),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
