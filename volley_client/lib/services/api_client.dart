import 'dart:convert';
import 'package:http/http.dart' as http;
import 'auth_service.dart';

/// API client that automatically handles token refresh on 401 errors
class ApiClient {
  final AuthService _authService;
  bool _isRefreshing = false;

  ApiClient(this._authService);

  /// Make an authenticated GET request with automatic token refresh
  Future<http.Response> get(
    String url, {
    Map<String, String>? headers,
  }) async {
    return _makeRequestWithRetry(
      () async {
        final token = await _authService.getToken();
        final requestHeaders = {
          'Content-Type': 'application/json',
          'X-Client-Type': 'mobile',
          if (token != null) 'Authorization': 'Bearer $token',
          ...?headers,
        };
        return http.get(Uri.parse(url), headers: requestHeaders);
      },
    );
  }

  /// Make an authenticated POST request with automatic token refresh
  Future<http.Response> post(
    String url, {
    Map<String, String>? headers,
    Object? body,
  }) async {
    return _makeRequestWithRetry(
      () async {
        final token = await _authService.getToken();
        final requestHeaders = {
          'Content-Type': 'application/json',
          'X-Client-Type': 'mobile',
          if (token != null) 'Authorization': 'Bearer $token',
          ...?headers,
        };
        return http.post(
          Uri.parse(url),
          headers: requestHeaders,
          body: body != null ? json.encode(body) : null,
        );
      },
    );
  }

  /// Make an authenticated PATCH request with automatic token refresh
  Future<http.Response> patch(
    String url, {
    Map<String, String>? headers,
    Object? body,
  }) async {
    return _makeRequestWithRetry(
      () async {
        final token = await _authService.getToken();
        final requestHeaders = {
          'Content-Type': 'application/json',
          'X-Client-Type': 'mobile',
          if (token != null) 'Authorization': 'Bearer $token',
          ...?headers,
        };
        return http.patch(
          Uri.parse(url),
          headers: requestHeaders,
          body: body != null ? json.encode(body) : null,
        );
      },
    );
  }

  /// Make an authenticated DELETE request with automatic token refresh
  Future<http.Response> delete(
    String url, {
    Map<String, String>? headers,
  }) async {
    return _makeRequestWithRetry(
      () async {
        final token = await _authService.getToken();
        final requestHeaders = {
          'Content-Type': 'application/json',
          'X-Client-Type': 'mobile',
          if (token != null) 'Authorization': 'Bearer $token',
          ...?headers,
        };
        return http.delete(Uri.parse(url), headers: requestHeaders);
      },
    );
  }

  /// Internal method that makes a request and retries once with token refresh on 401
  Future<http.Response> _makeRequestWithRetry(
    Future<http.Response> Function() makeRequest,
  ) async {
    // First attempt
    http.Response response = await makeRequest();

    // If 401 and not already refreshing, try to refresh token and retry once
    if (response.statusCode == 401 && !_isRefreshing) {
      _isRefreshing = true;
      try {
        // Attempt to refresh the access token
        await _authService.refreshAccessToken();

        // Retry the request with new token
        response = await makeRequest();
      } catch (e) {
        // Refresh failed, user needs to login again
        // The error will be handled by the calling code
        rethrow;
      } finally {
        _isRefreshing = false;
      }
    }

    return response;
  }
}
