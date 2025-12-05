# Volley Client

Flutter mobile and web application for organizing and joining pickup sports games.

## Overview

Volley is a mobile app and web app that allows users to:
- Browse and join pickup sports games
- Organize and manage their own games
- Track rosters and waitlists
- Handle payments for games

The app supports iOS, Android, and web platforms.

## Getting Started

### Prerequisites

- Flutter SDK (3.10.1 or higher)
- Dart SDK
- For iOS development: Xcode
- For Android development: Android Studio

### Installation

1. Install dependencies:
```bash
flutter pub get
```

2. Run the app:
```bash
# For development on your preferred platform
flutter run

# For web
flutter run -d chrome

# For iOS
flutter run -d ios

# For Android
flutter run -d android
```

### Running Tests

```bash
flutter test
```

### Project Structure

```
lib/
├── main.dart           # App entry point
├── screens/            # Screen widgets
│   └── home_screen.dart
├── widgets/            # Reusable widgets
├── models/             # Data models
├── services/           # API and business logic
│   └── api_service.dart
└── providers/          # State management providers
```

## Dependencies

Key packages used in this project:
- `http` - HTTP client for API calls
- `provider` - State management
- `go_router` - Navigation and routing
- `shared_preferences` - Local storage
- `intl` - Date and time utilities

## Backend

The app connects to a Go/Gin REST API backend. Update the API base URL in `lib/services/api_service.dart` to point to your backend server.

## Development

This project follows Flutter best practices and Material Design 3 guidelines.
