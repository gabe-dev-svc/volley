import 'package:flutter/foundation.dart';
import 'package:geolocator/geolocator.dart';
import 'package:geocoding/geocoding.dart' as geocoding;
import 'package:volley_client/providers/auth_provider.dart';
import '../models/game.dart';
import '../services/game_service.dart';

class GameProvider extends ChangeNotifier {
  final GameService _gameService = GameService();
  final AuthProvider _authProvider;

  GameProvider(this._authProvider);

  List<GameSummary> _games = [];
  bool _isLoading = false;
  String? _errorMessage;

  // Location state
  double? _latitude;
  double? _longitude;
  double _radius = 16093.4; // 10 miles in meters

  // Filter state
  GameCategory? _selectedCategory;

  List<GameSummary> get games => _games;
  bool get isLoading => _isLoading;
  String? get errorMessage => _errorMessage;
  double? get latitude => _latitude;
  double? get longitude => _longitude;
  double get radius => _radius;
  GameCategory? get selectedCategory => _selectedCategory;

  // Get radius in miles for display
  double get radiusInMiles => _radius / 1609.34;

  // Set location
  void setLocation(double lat, double lon) {
    _latitude = lat;
    _longitude = lon;
    notifyListeners();
  }

  // Geocode location (zip code or address)
  Future<bool> geocodeLocation(String location) async {
    try {
      _isLoading = true;
      _errorMessage = null;
      notifyListeners();

      // Fallback coordinates for common zip codes (for testing)
      final Map<String, Map<String, double>> zipCodeFallbacks = {
        '90210': {'lat': 34.0901, 'lng': -118.4065}, // Beverly Hills
        '10001': {'lat': 40.7506, 'lng': -73.9971},  // NYC
        '77001': {'lat': 29.7604, 'lng': -95.3698},  // Houston
        '60601': {'lat': 41.8858, 'lng': -87.6229},  // Chicago
        '94102': {'lat': 37.7796, 'lng': -122.4192}, // San Francisco
      };

      // Check if it's a known zip code
      if (zipCodeFallbacks.containsKey(location)) {
        _latitude = zipCodeFallbacks[location]!['lat'];
        _longitude = zipCodeFallbacks[location]!['lng'];
        _isLoading = false;
        notifyListeners();
        return true;
      }

      // Try geocoding with the API
      String searchLocation = location;
      if (RegExp(r'^\d{5}$').hasMatch(location)) {
        searchLocation = '$location, USA';
      }

      List<geocoding.Location>? locations;
      try {
        locations = await geocoding.locationFromAddress(searchLocation);
      } catch (e) {
        // If geocoding fails, try without USA suffix
        if (searchLocation.contains('USA')) {
          try {
            locations = await geocoding.locationFromAddress(location);
          } catch (_) {
            // Silently catch and continue to error handling below
          }
        }
      }

      if (locations != null && locations.isNotEmpty) {
        _latitude = locations.first.latitude;
        _longitude = locations.first.longitude;
        _isLoading = false;
        notifyListeners();
        return true;
      } else {
        _errorMessage = 'Location not found. Try: 90210, 10001, 77001, 60601, or 94102';
        _isLoading = false;
        notifyListeners();
        return false;
      }
    } catch (e) {
      _errorMessage = 'Geocoding unavailable. Try: 90210, 10001, 77001, 60601, or 94102';
      _isLoading = false;
      notifyListeners();
      return false;
    }
  }

  // Set radius in miles
  void setRadiusInMiles(double miles) {
    _radius = miles * 1609.34; // Convert to meters
    notifyListeners();
  }

  // Set category filter
  void setCategory(GameCategory? category) {
    _selectedCategory = category;
    notifyListeners();
    // Reload games with new filter
    if (_latitude != null && _longitude != null) {
      loadGames();
    }
  }

  // Get current location
  Future<bool> getCurrentLocation() async {
    try {
      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
        if (permission == LocationPermission.denied) {
          _errorMessage = 'Location permissions denied. Enter a zip code: 90210, 10001, 77001, 60601, or 94102';
          notifyListeners();
          return false;
        }
      }

      if (permission == LocationPermission.deniedForever) {
        _errorMessage = 'Location permissions permanently denied. Enter a zip code: 90210, 10001, 77001, 60601, or 94102';
        notifyListeners();
        return false;
      }

      final position = await Geolocator.getCurrentPosition();
      _latitude = position.latitude;
      _longitude = position.longitude;
      notifyListeners();
      return true;
    } catch (e) {
      _errorMessage = 'Failed to get location. Enter a zip code: 90210, 10001, 77001, 60601, or 94102';
      notifyListeners();
      return false;
    }
  }

  // Load games based on current filters
  Future<void> loadGames() async {
    if (_latitude == null || _longitude == null) {
      _errorMessage = 'Location not set';
      notifyListeners();
      return;
    }

    _isLoading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      final authToken = await this._authProvider.getToken();

      // If no category selected, fetch all categories
      final categories = _selectedCategory != null
          ? [_selectedCategory!]
          : [
              GameCategory.soccer,
              GameCategory.basketball,
              GameCategory.pickleball,
              GameCategory.flagFootball,
              GameCategory.volleyball,
            ];

      _games = await _gameService.listGames(
        categories: categories,
        latitude: _latitude!,
        longitude: _longitude!,
        radius: _radius,
        authToken: authToken ?? ""
      );

      _isLoading = false;
      notifyListeners();
    } catch (e) {
      _errorMessage = e.toString().replaceAll('Exception: ', '');
      _isLoading = false;
      notifyListeners();
    }
  }

  // Join a game
  Future<bool> joinGame(String gameId, String authToken) async {
    try {
      await _gameService.joinGame(gameId, authToken);
      // Reload the games list to update participant count
      await loadGames();
      return true;
    } catch (e) {
      _errorMessage = e.toString().replaceAll('Exception: ', '');
      notifyListeners();
      return false;
    }
  }

  // Leave a game
  Future<bool> leaveGame(String gameId, String authToken) async {
    try {
      await _gameService.leaveGame(gameId, authToken);
      // Reload the games list to update participant count
      await loadGames();
      return true;
    } catch (e) {
      _errorMessage = e.toString().replaceAll('Exception: ', '');
      notifyListeners();
      return false;
    }
  }

  // Clear error message
  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }
}
