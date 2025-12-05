import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/place.dart';

class LocationService {
  static const String baseUrl = 'http://localhost:8080/v1';

  Future<List<PlacePrediction>> autocomplete(
    String input,
    String authToken, {
    double? latitude,
    double? longitude,
  }) async {
    // Build request body for new Google Places API v1
    final requestBody = {
      'textQuery': input,
    };

    final response = await http.post(
      Uri.parse('$baseUrl/places/search'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $authToken',
      },
      body: json.encode(requestBody),
    );

    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      final predictions = data['predictions'] as List?;

      if (predictions == null || predictions.isEmpty) {
        return [];
      }

      return predictions
          .map((json) => PlacePrediction.fromJson(json as Map<String, dynamic>))
          .toList();
    } else {
      final data = json.decode(response.body);
      throw Exception(data['error'] ?? 'Failed to fetch location suggestions');
    }
  }

  Future<PlaceDetails> getPlaceDetails(
    String placeId,
    String authToken,
  ) async {
    final response = await http.get(
      Uri.parse('$baseUrl/places/$placeId'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $authToken',
      },
    );

    if (response.statusCode == 200) {
      return PlaceDetails.fromJson(json.decode(response.body));
    } else {
      final data = json.decode(response.body);
      throw Exception(data['error'] ?? 'Failed to fetch place details');
    }
  }
}
