import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';

import 'package:volley_client/screens/login_screen.dart';
import 'package:volley_client/screens/signup_screen.dart';
import 'package:volley_client/providers/auth_provider.dart';

void main() {
  testWidgets('Login screen renders correctly', (WidgetTester tester) async {
    await tester.pumpWidget(
      ChangeNotifierProvider(
        create: (_) => AuthProvider(),
        child: const MaterialApp(
          home: LoginScreen(),
        ),
      ),
    );

    // Should show login screen content
    expect(find.text('Ready to Play?'), findsOneWidget);
    expect(find.text('Log in to join the game.'), findsOneWidget);

    // Check for email and password fields
    expect(find.text('Email'), findsOneWidget);
    expect(find.text('Password'), findsOneWidget);
    expect(find.text('Forgot Password?'), findsOneWidget);

    // Check for login button
    expect(find.widgetWithText(ElevatedButton, 'Log In'), findsOneWidget);

    // Check for social login buttons
    expect(find.text('Continue with Google'), findsOneWidget);
    expect(find.text('Continue with Apple'), findsOneWidget);
    expect(find.text('Continue with Facebook'), findsOneWidget);

    // Check for sign up link
    expect(find.text('Sign Up'), findsOneWidget);
  });

  testWidgets('Signup screen renders correctly', (WidgetTester tester) async {
    await tester.pumpWidget(
      ChangeNotifierProvider(
        create: (_) => AuthProvider(),
        child: const MaterialApp(
          home: SignupScreen(),
        ),
      ),
    );

    // Should show signup screen content
    expect(find.text('Create Account'), findsOneWidget);
    expect(find.text('Sign up to start playing.'), findsOneWidget);

    // Check for all required fields
    expect(find.text('First Name'), findsOneWidget);
    expect(find.text('Last Name'), findsOneWidget);
    expect(find.text('Email'), findsOneWidget);
    expect(find.text('Password'), findsOneWidget);
    expect(find.text('Confirm Password'), findsOneWidget);

    // Check for signup button
    expect(find.widgetWithText(ElevatedButton, 'Sign Up'), findsOneWidget);
  });

  testWidgets('Login form validates empty fields', (WidgetTester tester) async {
    await tester.pumpWidget(
      ChangeNotifierProvider(
        create: (_) => AuthProvider(),
        child: const MaterialApp(
          home: LoginScreen(),
        ),
      ),
    );

    // Tap login button without entering any data
    await tester.tap(find.widgetWithText(ElevatedButton, 'Log In'));
    await tester.pump();

    // Should show validation errors
    expect(find.text('Please enter your email'), findsOneWidget);
    expect(find.text('Please enter your password'), findsOneWidget);
  });

  testWidgets('Signup form validates empty fields',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      ChangeNotifierProvider(
        create: (_) => AuthProvider(),
        child: const MaterialApp(
          home: SignupScreen(),
        ),
      ),
    );

    // Scroll down to find the button
    await tester.drag(find.byType(SingleChildScrollView), const Offset(0, -300));
    await tester.pumpAndSettle();

    // Tap signup button without entering any data
    await tester.tap(find.widgetWithText(ElevatedButton, 'Sign Up'));
    await tester.pump();

    // Should show validation errors
    expect(find.text('Please enter your first name'), findsOneWidget);
    expect(find.text('Please enter your last name'), findsOneWidget);
    expect(find.text('Please enter your email'), findsOneWidget);
    expect(find.text('Please enter your password'), findsOneWidget);
    expect(find.text('Please confirm your password'), findsOneWidget);
  });

  testWidgets('Signup form validates password length',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      ChangeNotifierProvider(
        create: (_) => AuthProvider(),
        child: const MaterialApp(
          home: SignupScreen(),
        ),
      ),
    );

    // Fill in all fields except use short password
    await tester.enterText(
      find.widgetWithText(TextFormField, 'Enter your first name'),
      'John',
    );
    await tester.enterText(
      find.widgetWithText(TextFormField, 'Enter your last name'),
      'Doe',
    );
    await tester.enterText(
      find.widgetWithText(TextFormField, 'Enter your email'),
      'john@example.com',
    );
    await tester.enterText(
      find.widgetWithText(TextFormField, 'Enter your password'),
      'short',
    );
    await tester.enterText(
      find.widgetWithText(TextFormField, 'Confirm your password'),
      'short',
    );

    // Scroll down to find the button
    await tester.drag(find.byType(SingleChildScrollView), const Offset(0, -300));
    await tester.pumpAndSettle();

    // Tap signup button
    await tester.tap(find.widgetWithText(ElevatedButton, 'Sign Up'));
    await tester.pump();

    // Should show password length error
    expect(find.text('Password must be at least 8 characters'), findsOneWidget);
  });
}
