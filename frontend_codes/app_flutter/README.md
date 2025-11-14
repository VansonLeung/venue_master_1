# Venue Master Mobile App

A Flutter mobile application for Venue Master - facility booking and management platform.

## Getting Started

### Prerequisites

- Flutter SDK 3.2.0 or higher
- Dart 3.2.0 or higher
- iOS: Xcode 14+ (for iOS development)
- Android: Android Studio (for Android development)

### Installation

1. **Install dependencies:**
   ```bash
   flutter pub get
   ```

2. **Generate code:**
   ```bash
   flutter pub run build_runner build --delete-conflicting-outputs
   ```

3. **Run the app:**
   ```bash
   # iOS
   flutter run -d ios

   # Android
   flutter run -d android

   # Web
   flutter run -d chrome
   ```

## Project Structure

```
lib/
├── core/              # Core utilities, config, theme
├── data/              # Data layer (models, repositories, providers)
├── features/          # Feature modules
└── widgets/           # Reusable widgets
```

## Key Features

- 🔐 Authentication (Login/Logout)
- 🏢 Facility browsing and booking
- 📅 Schedule viewing
- 🍔 Food ordering
- 🅿️ Parking reservations
- 🛍️ Pro shop
- 👤 User profile management
- 🔔 Notifications

## Architecture

- **State Management**: Riverpod
- **Navigation**: GoRouter
- **API**: GraphQL + REST (Dio)
- **Storage**: Shared Preferences + Secure Storage

## Development

### Code Generation

Watch mode (automatic rebuild on file changes):
```bash
flutter pub run build_runner watch
```

One-time build:
```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

### Testing

```bash
# Run all tests
flutter test

# Run with coverage
flutter test --coverage

# Run specific test
flutter test test/unit/auth_test.dart
```

### Code Quality

```bash
# Analyze code
flutter analyze

# Format code
dart format .
```

## Backend Integration

The app connects to the Venue Master backend services:

- **Gateway**: http://localhost:8080
- **Auth Service**: http://localhost:8081
- **Booking Service**: http://localhost:8083
- **Food Service**: http://localhost:8084
- **Parking Service**: http://localhost:8085
- **Shop Service**: http://localhost:8086

**Note**: For Android Emulator, use `http://10.0.2.2` instead of `localhost`.

## Environment Configuration

Create a `.env` file for environment-specific configuration:

```env
API_BASE_URL=http://localhost
GRAPHQL_ENDPOINT=http://localhost:8080/graphql
```

## Contributing

1. Create a feature branch
2. Make your changes
3. Write tests
4. Submit a pull request

## License

Proprietary - Venue Master
