import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/game.dart';

class GameService {
  static const String baseUrl = 'http://localhost:8080/v1';

  Future<List<GameSummary>> listGames({
    required List<GameCategory> categories,
    required double latitude,
    required double longitude,
    double radius = 16093.4, // 10 miles in meters
    String timeFilter = 'upcoming',
    GameStatus? status,
    int limit = 20,
    int offset = 0,
    String authToken = "",
  }) async {
    final queryParams = {
      'categories': categories.map((c) => c.toApiString()).join(','),
      'latitude': latitude.toString(),
      'longitude': longitude.toString(),
      'radius': radius.toString(),
      'timeFilter': timeFilter,
      'limit': limit.toString(),
      'offset': offset.toString(),
    };

    if (status != null) {
      queryParams['status'] = status.toString().split('.').last;
    }

    final uri = Uri.parse('$baseUrl/games').replace(queryParameters: queryParams);

    Map<String,String> headers = { 'Content-Type': 'application/json' };
    if (authToken != "") {
      headers["Authorization"] = 'Bearer $authToken';
    }

    final response = await http.get(
      uri,
      headers: headers,
    );

    if (response.statusCode == 200) {
      final data = json.decode(response.body);
      final gamesJson = data['games'] as List;
      return gamesJson.map((json) => GameSummary.fromJson(json as Map<String, dynamic>)).toList();
    } else {
      throw Exception('Failed to load games: ${response.statusCode}');
    }
  }

  Future<void> joinGame(String gameId, String authToken) async {
    final response = await http.post(
      Uri.parse('$baseUrl/games/$gameId/participation'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $authToken',
      },
    );

    if (response.statusCode != 200) {
      final data = json.decode(response.body);
      throw Exception(data['error'] ?? 'Failed to join game');
    }
  }

  Future<void> leaveGame(String gameId, String authToken) async {
    final response = await http.delete(
      Uri.parse('$baseUrl/games/$gameId/participation'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $authToken',
      },
    );

    if (response.statusCode != 200) {
      final data = json.decode(response.body);
      throw Exception(data['error'] ?? 'Failed to leave game');
    }
  }

  Future<Game> getGame(String gameId, {String? authToken}) async {
    Map<String, String> headers = {'Content-Type': 'application/json'};
    if (authToken != null && authToken.isNotEmpty) {
      headers['Authorization'] = 'Bearer $authToken';
    }

    final response = await http.get(
      Uri.parse('$baseUrl/games/$gameId'),
      headers: headers,
    );

    if (response.statusCode == 200) {
      return Game.fromJson(json.decode(response.body));
    } else {
      throw Exception('Failed to load game: ${response.statusCode}');
    }
  }

  Future<Game> createGame(CreateGameRequest request, String authToken) async {
    final response = await http.post(
      Uri.parse('$baseUrl/games'),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $authToken',
        'X-Client-Type': 'mobile',
      },
      body: json.encode(request.toJson()),
    );

    if (response.statusCode == 201) {
      return Game.fromJson(json.decode(response.body));
    } else {
      final data = json.decode(response.body);
      throw Exception(data['error'] ?? 'Failed to create game: ${response.statusCode}');
    }
  }
}
