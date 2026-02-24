import 'package:flutter/material.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

class ScanScreen extends StatefulWidget {
  const ScanScreen({super.key});

  @override
  State<ScanScreen> createState() => _ScanScreenState();
}

class _ScanScreenState extends State<ScanScreen> {
  String? lastCode;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Scan QR Code'),
      ),
      body: Column(
        children: [
          Expanded(
            child: MobileScanner(
              onDetect: (capture) {
                final code = capture.barcodes.firstOrNull;
                if (code == null) return;
                setState(() {
                  lastCode = code.rawValue;
                });
              },
            ),
          ),
          Container(
            padding: const EdgeInsets.all(16),
            color: const Color(0xFFF1F5F9),
            child: Row(
              children: [
                const Icon(Icons.qr_code_2, color: Color(0xFF2563EB)),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    lastCode ?? 'Point the camera at a QR code',
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
