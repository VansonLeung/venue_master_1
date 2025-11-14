#!/bin/bash
# Venue Master Flutter App Setup Script

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MOBILE_DIR="$PROJECT_ROOT/mobile"

echo "========================================="
echo "  Venue Master Flutter Setup"
echo "========================================="
echo ""

# Check if Flutter is installed
if ! command -v flutter &> /dev/null; then
    echo "❌ Flutter is not installed!"
    echo ""
    echo "Please install Flutter first:"
    echo "  macOS: brew install --cask flutter"
    echo "  Or follow: https://docs.flutter.dev/get-started/install"
    echo ""
    exit 1
fi

echo "✓ Flutter is installed"
flutter --version
echo ""

# Check Flutter setup
echo "Running flutter doctor..."
flutter doctor
echo ""

# Create Flutter project if it doesn't exist
if [ ! -d "$MOBILE_DIR" ]; then
    echo "Creating Flutter project..."
    cd "$PROJECT_ROOT"
    flutter create mobile \
        --org com.venuemaster \
        --project-name venue_master_mobile \
        --platforms ios,android,web \
        --description "Venue Master - Facility booking and management app"

    echo "✓ Flutter project created at $MOBILE_DIR"
else
    echo "✓ Flutter project already exists"
fi

cd "$MOBILE_DIR"

# Create folder structure
echo ""
echo "Creating project structure..."

mkdir -p lib/core/config
mkdir -p lib/core/constants
mkdir -p lib/core/theme
mkdir -p lib/core/utils
mkdir -p lib/data/models
mkdir -p lib/data/providers
mkdir -p lib/data/repositories
mkdir -p lib/features/auth/screens
mkdir -p lib/features/auth/widgets
mkdir -p lib/features/home/screens
mkdir -p lib/features/home/widgets
mkdir -p lib/features/bookings/screens
mkdir -p lib/features/bookings/widgets
mkdir -p lib/features/facilities/screens
mkdir -p lib/features/facilities/widgets
mkdir -p lib/features/food/screens
mkdir -p lib/features/food/widgets
mkdir -p lib/features/parking/screens
mkdir -p lib/features/parking/widgets
mkdir -p lib/features/shop/screens
mkdir -p lib/features/shop/widgets
mkdir -p lib/features/profile/screens
mkdir -p lib/features/profile/widgets
mkdir -p lib/features/notifications/screens
mkdir -p lib/features/notifications/widgets
mkdir -p lib/widgets/common
mkdir -p lib/widgets/layouts
mkdir -p assets/images
mkdir -p assets/icons
mkdir -p assets/fonts
mkdir -p test/unit
mkdir -p test/widget
mkdir -p test/integration

echo "✓ Project structure created"

# Backup original pubspec.yaml
if [ -f pubspec.yaml ]; then
    cp pubspec.yaml pubspec.yaml.backup
    echo "✓ Backed up pubspec.yaml"
fi

echo ""
echo "========================================="
echo "  Setup Complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "  1. cd mobile"
echo "  2. Review and update pubspec.yaml with dependencies"
echo "  3. Run: flutter pub get"
echo "  4. Run: flutter run"
echo ""
echo "See FLUTTER_SETUP_GUIDE.md for detailed instructions"
echo ""
