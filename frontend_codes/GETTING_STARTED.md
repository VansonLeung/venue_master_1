# Getting Started with Venue Master Flutter App

## Prerequisites

Before you begin, ensure you have the following installed:

### 1. Install Flutter

**macOS (recommended):**
```bash
# Using Homebrew
brew install --cask flutter

# Or download directly
# Visit: https://docs.flutter.dev/get-started/install/macos
```

**After installation, add Flutter to your PATH:**
```bash
# Add to ~/.zshrc or ~/.bash_profile
export PATH="$PATH:$HOME/development/flutter/bin"

# Apply changes
source ~/.zshrc  # or source ~/.bash_profile
```

### 2. Verify Flutter Installation

```bash
flutter doctor
```

This will check your environment and display a report. Fix any issues it identifies.

### 3. Install Additional Tools

**For iOS development:**
```bash
# Install Xcode from App Store
# Then accept the license:
sudo xcodebuild -license accept

# Install CocoaPods
sudo gem install cocoapods
```

**For Android development:**
- Download Android Studio from https://developer.android.com/studio
- Install Android SDK and emulator through Android Studio

## Project Setup

### Step 1: Navigate to the Flutter Project

```bash
cd /Users/van/Downloads/venue_master/frontend_codes/app_flutter
```

### Step 2: Install Dependencies

```bash
flutter pub get
```

This will download all the packages specified in `pubspec.yaml`.

### Step 3: Generate Code

The project uses code generation for models and providers:

```bash
# One-time build
flutter pub run build_runner build --delete-conflicting-outputs

# Or use watch mode (recommended during development)
flutter pub run build_runner watch
```

**Note:** The first code generation will fail because we need to create `.g.dart` files. This is normal for a new project. We'll add implementations progressively.

### Step 4: Run the App

**iOS Simulator:**
```bash
# List available devices
flutter devices

# Run on iOS
flutter run -d ios
```

**Android Emulator:**
```bash
# Make sure an emulator is running
# Then run:
flutter run -d android
```

**Chrome (Web):**
```bash
flutter run -d chrome
```

## Project Structure

```
frontend_codes/app_flutter/
├── lib/
│   ├── core/                    # Core functionality
│   │   ├── config/             # API & app configuration
│   │   ├── constants/          # App constants
│   │   ├── theme/              # Theme & styling
│   │   └── utils/              # Utility functions
│   ├── data/                    # Data layer
│   │   ├── models/             # Data models
│   │   ├── providers/          # API providers
│   │   └── repositories/       # Data repositories
│   ├── features/                # Feature modules
│   │   ├── auth/               # Authentication
│   │   ├── home/               # Home dashboard
│   │   ├── bookings/           # Bookings management
│   │   ├── facilities/         # Facility browsing
│   │   └── ...                 # Other features
│   ├── widgets/                 # Reusable widgets
│   └── main.dart               # App entry point
├── assets/                      # Images, icons, fonts
├── test/                        # Tests
└── pubspec.yaml                # Dependencies
```

## Development Workflow

### Hot Reload

While the app is running, make changes to your code and press:
- `r` - Hot reload
- `R` - Hot restart
- `q` - Quit

### Code Generation

When you modify models or add new providers:

```bash
# In watch mode (runs automatically)
flutter pub run build_runner watch

# Or build once
flutter pub run build_runner build --delete-conflicting-outputs
```

### Running Tests

```bash
# Run all tests
flutter test

# Run specific test file
flutter test test/unit/auth_test.dart

# Run with coverage
flutter test --coverage
```

### Code Quality

```bash
# Analyze code
flutter analyze

# Format code
dart format .

# Fix common issues
dart fix --apply
```

## Connecting to Backend

### Local Development

The app is configured to connect to your local backend services:

**For iOS Simulator:**
- Uses `http://localhost:8080` (already configured)

**For Android Emulator:**
- Update `lib/core/config/api_config.dart`
- Change `baseUrl` to use `_baseUrlAndroid` (http://10.0.2.2)

### Start Backend Services

Before running the app, ensure your backend is running:

```bash
# In the project root
cd codes
docker-compose up -d

# Verify services are running
./scripts/test-api.sh
```

## Common Issues & Solutions

### Issue 1: "Flutter not found"
```bash
# Make sure Flutter is in your PATH
export PATH="$PATH:$HOME/development/flutter/bin"
source ~/.zshrc
```

### Issue 2: "Unable to find git in your PATH"
```bash
# Install Xcode Command Line Tools
xcode-select --install
```

### Issue 3: "CocoaPods not installed"
```bash
sudo gem install cocoapods
```

### Issue 4: "Build failed - .g.dart files missing"
```bash
# Generate the files
flutter pub run build_runner build --delete-conflicting-outputs
```

### Issue 5: "Cannot connect to backend"
```bash
# Check if backend is running
curl http://localhost:8080/healthz

# For Android, use 10.0.2.2 instead of localhost in api_config.dart
```

## Next Steps

### Phase 1: Complete Authentication
1. Implement API client for login
2. Add secure storage for tokens
3. Create auth state management
4. Add logout functionality

### Phase 2: Build Home Screen
1. Create home dashboard UI
2. Fetch user data
3. Display upcoming bookings
4. Add quick action buttons

### Phase 3: Implement Bookings
1. Create facility list screen
2. Add facility details
3. Implement booking flow
4. Add booking history

## Useful Commands Reference

```bash
# Project management
flutter clean              # Clean build artifacts
flutter pub get           # Get dependencies
flutter pub upgrade       # Upgrade dependencies

# Development
flutter run               # Run app
flutter run -v           # Run with verbose logging
flutter run --release    # Run in release mode

# Code generation
flutter pub run build_runner build --delete-conflicting-outputs
flutter pub run build_runner watch

# Testing
flutter test                    # Run all tests
flutter test --coverage        # With coverage report
flutter test test/unit/        # Run unit tests only

# Code quality
flutter analyze           # Static analysis
dart format .            # Format code
dart fix --apply         # Auto-fix issues

# Build
flutter build apk        # Build Android APK
flutter build ios        # Build iOS app
flutter build web        # Build web app
```

## Resources

- **Flutter Documentation**: https://docs.flutter.dev/
- **Riverpod Docs**: https://riverpod.dev/
- **Material Design 3**: https://m3.material.io/
- **Venue Master API Docs**: See `/docs/` in project root

## Need Help?

- Check the main `FLUTTER_SETUP_GUIDE.md`
- Review `FLUTTER_IMPLEMENTATION_ROADMAP.md`
- Run `flutter doctor` to diagnose issues
- Check Flutter community resources

---

**Happy Coding! 🚀**
