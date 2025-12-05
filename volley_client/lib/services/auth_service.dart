import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:sign_in_with_apple/sign_in_with_apple.dart';
import 'package:flutter_facebook_auth/flutter_facebook_auth.dart';
import '../models/auth_response.dart';
import '../models/user.dart';

class AuthService {
  static const String baseUrl = 'http://localhost:8080/v1';
  final FlutterSecureStorage _secureStorage = const FlutterSecureStorage();
  final GoogleSignIn _googleSignIn = GoogleSignIn();

  // Keys for secure storage
  static const String _tokenKey = 'auth_token';
  static const String _userKey = 'user_data';

  // Login with email and password
  Future<AuthResponse> login(String email, String password) async {
    final response = await http.post(
      Uri.parse('$baseUrl/auth/login'),
      headers: {
        'Content-Type': 'application/json',
        'X-Client-Type': 'mobile',
      },
      body: json.encode({
        'email': email,
        'password': password,
      }),
    );

    if (response.statusCode == 200) {
      final authResponse = AuthResponse.fromJson(json.decode(response.body));
      await _saveAuthData(authResponse);
      return authResponse;
    } else if (response.statusCode == 401) {
      throw Exception('Invalid credentials');
    } else {
      final error = json.decode(response.body);
      throw Exception(error['error'] ?? 'Login failed');
    }
  }

  // Register new user
  Future<AuthResponse> register({
    required String email,
    required String password,
    required String firstName,
    required String lastName,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl/auth/register'),
      headers: {
        'Content-Type': 'application/json',
        'X-Client-Type': 'mobile',
      },
      body: json.encode({
        'email': email,
        'password': password,
        'firstName': firstName,
        'lastName': lastName,
      }),
    );

    if (response.statusCode == 201) {
      final authResponse = AuthResponse.fromJson(json.decode(response.body));
      await _saveAuthData(authResponse);
      return authResponse;
    } else if (response.statusCode == 409) {
      throw Exception('User already exists');
    } else {
      final error = json.decode(response.body);
      throw Exception(error['error'] ?? 'Registration failed');
    }
  }

  // Google Sign In
  Future<AuthResponse> signInWithGoogle() async {
    try {
      final GoogleSignInAccount? googleUser = await _googleSignIn.signIn();
      if (googleUser == null) {
        throw Exception('Google sign in cancelled');
      }

      // ignore: unused_local_variable
      final GoogleSignInAuthentication googleAuth =
          await googleUser.authentication;

      // TODO: Send Google token to backend for verification and user creation
      // This would require a new endpoint in the API like /auth/google
      // Use googleAuth.accessToken and googleAuth.idToken
      throw UnimplementedError(
          'Google OAuth backend endpoint not yet implemented');
    } catch (e) {
      throw Exception('Google sign in failed: $e');
    }
  }

  // Apple Sign In
  Future<AuthResponse> signInWithApple() async {
    try {
      // ignore: unused_local_variable
      final credential = await SignInWithApple.getAppleIDCredential(
        scopes: [
          AppleIDAuthorizationScopes.email,
          AppleIDAuthorizationScopes.fullName,
        ],
      );

      // TODO: Send Apple credential to backend for verification and user creation
      // This would require a new endpoint in the API like /auth/apple
      // Use credential.identityToken and credential.authorizationCode
      throw UnimplementedError(
          'Apple OAuth backend endpoint not yet implemented');
    } catch (e) {
      throw Exception('Apple sign in failed: $e');
    }
  }

  // Facebook Sign In
  Future<AuthResponse> signInWithFacebook() async {
    try {
      final LoginResult result = await FacebookAuth.instance.login();

      if (result.status == LoginStatus.success) {
        // ignore: unused_local_variable
        final AccessToken? accessToken = result.accessToken;

        // TODO: Send Facebook token to backend for verification and user creation
        // This would require a new endpoint in the API like /auth/facebook
        // Use accessToken.tokenString
        throw UnimplementedError(
            'Facebook OAuth backend endpoint not yet implemented');
      } else {
        throw Exception('Facebook sign in cancelled or failed');
      }
    } catch (e) {
      throw Exception('Facebook sign in failed: $e');
    }
  }

  // Save authentication data to secure storage
  Future<void> _saveAuthData(AuthResponse authResponse) async {
    if (authResponse.token != null) {
      await _secureStorage.write(key: _tokenKey, value: authResponse.token);
    }
    await _secureStorage.write(
      key: _userKey,
      value: json.encode(authResponse.user.toJson()),
    );
  }

  // Get stored token
  Future<String?> getToken() async {
    return await _secureStorage.read(key: _tokenKey);
  }

  // Get stored user
  Future<User?> getStoredUser() async {
    final userData = await _secureStorage.read(key: _userKey);
    if (userData != null) {
      return User.fromJson(json.decode(userData));
    }
    return null;
  }

  // Check if user is logged in
  Future<bool> isLoggedIn() async {
    final token = await getToken();
    return token != null;
  }

  // Logout
  Future<void> logout() async {
    await _secureStorage.delete(key: _tokenKey);
    await _secureStorage.delete(key: _userKey);
    await _googleSignIn.signOut();
    await FacebookAuth.instance.logOut();
  }
}
