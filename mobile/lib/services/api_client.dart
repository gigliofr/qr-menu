import 'dart:convert';
import 'package:http/http.dart' as http;

class ApiClient {
  ApiClient({this.baseUrl = 'http://localhost:8080/api/v1'});

  final String baseUrl;

  Future<List<dynamic>> getMenus(String token) async {
    final response = await http.get(
      Uri.parse('$baseUrl/menus'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );

    if (response.statusCode >= 200 && response.statusCode < 300) {
      return jsonDecode(response.body)['data'] as List<dynamic>;
    }

    throw Exception('Failed to fetch menus');
  }
}
