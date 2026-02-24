import 'package:flutter/material.dart';
import '../widgets/metric_card.dart';
import '../widgets/section_header.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('QR Menu Dashboard'),
        actions: [
          IconButton(
            icon: const Icon(Icons.qr_code_scanner),
            onPressed: () => Navigator.pushNamed(context, '/scan'),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const SectionHeader(title: 'Today'),
          const SizedBox(height: 8),
          const MetricCard(
            title: 'Menu Views',
            value: '4.2k',
            subtitle: '+18% vs last week',
            icon: Icons.visibility,
          ),
          const SizedBox(height: 12),
          const MetricCard(
            title: 'Active Menus',
            value: '6',
            subtitle: '2 drafts',
            icon: Icons.restaurant_menu,
          ),
          const SizedBox(height: 12),
          const MetricCard(
            title: 'Cache Hit Rate',
            value: '89%',
            subtitle: 'Avg 2ms response',
            icon: Icons.speed,
          ),
          const SizedBox(height: 24),
          SectionHeader(
            title: 'Quick Actions',
            action: TextButton(
              onPressed: () => Navigator.pushNamed(context, '/menus'),
              child: const Text('View Menus'),
            ),
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 12,
            runSpacing: 12,
            children: [
              _ActionCard(
                title: 'Create Menu',
                icon: Icons.add_circle_outline,
                onTap: () => Navigator.pushNamed(context, '/menus'),
              ),
              _ActionCard(
                title: 'Scan QR',
                icon: Icons.qr_code_scanner,
                onTap: () => Navigator.pushNamed(context, '/scan'),
              ),
              _ActionCard(
                title: 'Analytics',
                icon: Icons.bar_chart,
                onTap: () {},
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _ActionCard extends StatelessWidget {
  const _ActionCard({
    required this.title,
    required this.icon,
    required this.onTap,
  });

  final String title;
  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        width: 160,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: const Color(0xFFE2E8F0)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, color: const Color(0xFF2563EB)),
            const SizedBox(height: 12),
            Text(title, style: Theme.of(context).textTheme.titleSmall),
          ],
        ),
      ),
    );
  }
}
