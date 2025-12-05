import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import 'providers/auth_provider.dart';
import 'providers/game_provider.dart';
import 'screens/home_screen.dart';
import 'screens/login_screen.dart';
import 'screens/signup_screen.dart';
import 'screens/games_near_you_screen.dart';
import 'screens/create_game_screen.dart';
import 'screens/game_details_screen.dart';
import 'theme/app_theme.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const VolleyApp());
}

class VolleyApp extends StatelessWidget {
  const VolleyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AuthProvider()..initialize()),
        ChangeNotifierProxyProvider<AuthProvider, GameProvider>(
          create: (context) =>
              GameProvider(Provider.of<AuthProvider>(context, listen: false)),
          update: (_, authProvider, gameProvider) => gameProvider!,
        ),
      ],
      child: Consumer<AuthProvider>(
        builder: (context, authProvider, _) {
          return MaterialApp.router(
            title: 'Volley',
            theme: AppTheme.lightTheme,
            routerConfig: _createRouter(authProvider),
          );
        },
      ),
    );
  }
}

GoRouter _createRouter(AuthProvider authProvider) {
  return GoRouter(
    initialLocation: authProvider.isAuthenticated ? '/' : '/login',
    redirect: (context, state) {
      final isAuthenticated = authProvider.isAuthenticated;
      final isLoginPage = state.matchedLocation == '/login';
      final isSignupPage = state.matchedLocation == '/signup';

      // If not authenticated and not on login/signup pages, redirect to login
      if (!isAuthenticated && !isLoginPage && !isSignupPage) {
        return '/login';
      }

      // If authenticated and on login/signup pages, redirect to home
      if (isAuthenticated && (isLoginPage || isSignupPage)) {
        return '/';
      }

      return null;
    },
    routes: [
      GoRoute(path: '/login', builder: (context, state) => const LoginScreen()),
      GoRoute(
        path: '/signup',
        builder: (context, state) => const SignupScreen(),
      ),
      GoRoute(path: '/', builder: (context, state) => const HomeScreen()),
      GoRoute(
        path: '/games',
        builder: (context, state) => const GamesNearYouScreen(),
      ),
      GoRoute(
        path: '/create-game',
        builder: (context, state) => const CreateGameScreen(),
      ),
      GoRoute(
        path: '/game/:gameId',
        builder: (context, state) {
          final gameId = state.pathParameters['gameId']!;
          return GameDetailsScreen(gameId: gameId);
        },
      ),
    ],
  );
}
