# Venue Master Flutter App - Setup & Implementation Guide

## Overview

This guide covers the setup and implementation of the Venue Master mobile application using Flutter. The app will provide a modern, mobile-first interface for users to book facilities, view schedules, order food, and manage their venue memberships.

## Prerequisites

### 1. Install Flutter

**macOS Installation:**
```bash
# Download Flutter SDK
cd ~/development
git clone https://github.com/flutter/flutter.git -b stable

# Add Flutter to PATH (add to ~/.zshrc or ~/.bash_profile)
export PATH="$PATH:$HOME/development/flutter/bin"

# Verify installation
flutter doctor
```

**Alternative (using Homebrew):**
```bash
brew install --cask flutter
flutter doctor
```

### 2. Install Additional Tools

```bash
# Install Xcode (for iOS development)
# Download from App Store

# Accept Xcode license
sudo xcodebuild -license accept

# Install CocoaPods (for iOS dependencies)
sudo gem install cocoapods

# Install Android Studio (for Android development)
# Download from https://developer.android.com/studio

# Run flutter doctor to check setup
flutter doctor -v
```

## Project Structure

```
venue_master/
├── codes/                    # Backend services (existing)
├── mobile/                   # Flutter app (new)
│   ├── lib/
│   │   ├── main.dart
│   │   ├── core/
│   │   │   ├── config/      # App configuration
│   │   │   ├── constants/   # Constants & enums
│   │   │   ├── theme/       # App theme & styles
│   │   │   └── utils/       # Utility functions
│   │   ├── data/
│   │   │   ├── models/      # Data models
│   │   │   ├── providers/   # API providers (GraphQL & REST)
│   │   │   └── repositories/ # Data repositories
│   │   ├── features/
│   │   │   ├── auth/        # Authentication
│   │   │   ├── home/        # Home dashboard
│   │   │   ├── bookings/    # Facility bookings
│   │   │   ├── facilities/  # Facility browsing
│   │   │   ├── food/        # Food ordering
│   │   │   ├── parking/     # Parking reservations
│   │   │   ├── shop/        # Pro shop
│   │   │   ├── profile/     # User profile
│   │   │   └── notifications/ # Notifications
│   │   └── widgets/
│   │       ├── common/      # Reusable widgets
│   │       └── layouts/     # Layout components
│   ├── assets/
│   │   ├── images/
│   │   ├── icons/
│   │   └── fonts/
│   ├── test/
│   ├── pubspec.yaml
│   └── README.md
└── docs/
```

## Architecture

### State Management: Riverpod
- Clean, testable state management
- Type-safe and compile-time safe
- Great for reactive programming

### API Layer: GraphQL + REST
- **GraphQL Client**: For complex queries (facilities, bookings)
- **HTTP Client (Dio)**: For REST endpoints
- Automatic token refresh
- Error handling & retry logic

### Navigation: GoRouter
- Declarative routing
- Deep linking support
- Type-safe navigation

### Local Storage: Shared Preferences + Secure Storage
- Shared Preferences: User preferences
- Flutter Secure Storage: Auth tokens

## Key Features to Implement

### Phase 1: Foundation & Authentication
1. **App Setup**
   - Project initialization
   - Folder structure
   - Theme configuration (colors, typography)
   - Base widgets

2. **Authentication**
   - Login screen
   - JWT token management
   - Automatic token refresh
   - Secure storage for credentials

3. **API Integration**
   - GraphQL client setup
   - REST client setup (Dio)
   - Error handling
   - Network interceptors

### Phase 2: Core Features
4. **Home Dashboard**
   - Welcome screen
   - Quick actions
   - Upcoming bookings
   - Notifications badge

5. **Facility Browsing**
   - List all facilities
   - Filter by availability
   - Facility details
   - Schedule viewer

6. **Booking Management**
   - Create new booking
   - View booking details
   - Cancel booking
   - Booking history

### Phase 3: Additional Services
7. **Food Ordering**
   - Browse menu
   - Add to cart
   - Place order

8. **Parking Reservations**
   - View available spots
   - Reserve parking
   - View reservations

9. **Pro Shop**
   - Browse products
   - Shopping cart
   - Checkout

### Phase 4: User Experience
10. **Profile & Settings**
    - View profile
    - Membership details
    - App settings
    - Logout

11. **Notifications**
    - Push notifications
    - In-app notifications
    - Notification history

12. **Polish & Testing**
    - Loading states
    - Error handling
    - Offline support
    - Unit tests
    - Widget tests

## Dependencies

### Core Dependencies
```yaml
dependencies:
  flutter:
    sdk: flutter

  # State Management
  flutter_riverpod: ^2.4.9
  riverpod_annotation: ^2.3.3

  # Navigation
  go_router: ^13.0.0

  # API Clients
  graphql_flutter: ^5.1.2
  dio: ^5.4.0

  # Storage
  shared_preferences: ^2.2.2
  flutter_secure_storage: ^9.0.0

  # JSON Serialization
  json_annotation: ^4.8.1
  freezed_annotation: ^2.4.1

  # UI Components
  flutter_svg: ^2.0.9
  cached_network_image: ^3.3.1
  shimmer: ^3.0.0

  # Utilities
  intl: ^0.18.1
  logger: ^2.0.2

dev_dependencies:
  flutter_test:
    sdk: flutter

  # Code Generation
  build_runner: ^2.4.7
  json_serializable: ^6.7.1
  freezed: ^2.4.6
  riverpod_generator: ^2.3.9

  # Linting
  flutter_lints: ^3.0.1

  # Testing
  mockito: ^5.4.4
```

## API Endpoints Configuration

```dart
// lib/core/config/api_config.dart
class ApiConfig {
  static const String baseUrl = 'http://localhost';

  // Service Ports
  static const int gatewayPort = 8080;
  static const int authPort = 8081;
  static const int userPort = 8082;
  static const int bookingPort = 8083;
  static const int foodPort = 8084;
  static const int parkingPort = 8085;
  static const int shopPort = 8086;
  static const int paymentPort = 8087;
  static const int notificationPort = 8088;

  // GraphQL Endpoint
  static String get graphqlUrl => '$baseUrl:$gatewayPort/graphql';

  // REST Endpoints
  static String get authUrl => '$baseUrl:$authPort/v1/auth';
  static String get userUrl => '$baseUrl:$userPort/v1/users';
  static String get bookingUrl => '$baseUrl:$bookingPort/v1';
  static String get foodUrl => '$baseUrl:$foodPort/v1';
  static String get parkingUrl => '$baseUrl:$parkingPort/v1/parking';
  static String get shopUrl => '$baseUrl:$shopPort/v1';
  static String get paymentUrl => '$baseUrl:$paymentPort/v1/payments';
  static String get notificationUrl => '$baseUrl:$notificationPort/v1/notifications';
}
```

## Design System

### Color Palette
```dart
// Primary Colors
const primaryColor = Color(0xFF2196F3);      // Blue
const secondaryColor = Color(0xFF4CAF50);    // Green
const accentColor = Color(0xFFFF9800);       // Orange

// Neutral Colors
const backgroundColor = Color(0xFFF5F5F5);
const surfaceColor = Color(0xFFFFFFFF);
const errorColor = Color(0xFFF44336);

// Text Colors
const textPrimary = Color(0xFF212121);
const textSecondary = Color(0xFF757575);
const textHint = Color(0xFFBDBDBD);
```

### Typography
```dart
// Headings
heading1: 32px, Bold
heading2: 24px, SemiBold
heading3: 20px, Medium

// Body
bodyLarge: 16px, Regular
bodyMedium: 14px, Regular
bodySmall: 12px, Regular

// Captions
caption: 12px, Regular
```

## Getting Started

### Step 1: Create Flutter Project

```bash
cd /Users/van/Downloads/venue_master
flutter create mobile --org com.venuemaster --project-name venue_master_mobile
cd mobile
```

### Step 2: Update Dependencies

Edit `pubspec.yaml` and add the dependencies listed above.

```bash
flutter pub get
```

### Step 3: Generate Code

```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

### Step 4: Run the App

```bash
# iOS Simulator
flutter run -d ios

# Android Emulator
flutter run -d android

# Chrome (for web testing)
flutter run -d chrome
```

## Development Workflow

### 1. Feature Development Pattern

For each feature (e.g., bookings):

```
1. Create models (data/models/booking.dart)
2. Create repository (data/repositories/booking_repository.dart)
3. Create providers (data/providers/booking_provider.dart)
4. Create UI screens (features/bookings/screens/)
5. Create widgets (features/bookings/widgets/)
6. Add navigation routes
7. Write tests
```

### 2. Testing Strategy

```bash
# Unit tests
flutter test test/unit/

# Widget tests
flutter test test/widget/

# Integration tests
flutter test test/integration/

# Run all tests with coverage
flutter test --coverage
```

### 3. Code Generation

```bash
# Watch mode (automatic rebuild)
flutter pub run build_runner watch

# One-time build
flutter pub run build_runner build --delete-conflicting-outputs
```

## Next Steps

1. ✅ Install Flutter SDK
2. ✅ Run `flutter doctor` and fix any issues
3. ✅ Create the Flutter project
4. ✅ Set up folder structure
5. ✅ Add dependencies
6. ✅ Configure API endpoints
7. ✅ Implement authentication flow
8. ✅ Build home dashboard
9. ✅ Implement booking features
10. ✅ Add remaining features

## Resources

- **Flutter Docs**: https://docs.flutter.dev/
- **Riverpod**: https://riverpod.dev/
- **GraphQL Flutter**: https://pub.dev/packages/graphql_flutter
- **Go Router**: https://pub.dev/packages/go_router
- **Material Design**: https://m3.material.io/

## Notes

- Use `http://10.0.2.2` instead of `localhost` for Android Emulator
- Use `http://localhost` for iOS Simulator
- Consider environment-based configuration for dev/staging/prod
- Implement proper error handling and loading states
- Add analytics and crash reporting (Firebase)
