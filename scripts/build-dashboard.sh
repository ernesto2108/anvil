#!/usr/bin/env bash
# build-dashboard.sh — Builds the Anvil Dashboard Wails app.
#
# Workaround for wailsapp/wails#4265: Wails v2 "Generating bindings: permission denied"
# on macOS. We skip bindings generation and build the Go binary manually.
#
# Usage: ./scripts/build-dashboard.sh [--dev]
#   --dev: start in dev mode (live reload frontend, auto-rebuild Go)

set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
APP_NAME="Anvil Dashboard"
BINARY_NAME="anvil-full"
BUILD_DIR="$PROJECT_DIR/build/bin"
APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"
MACOS_DIR="$APP_BUNDLE/Contents/MacOS"
FRONTEND_DIR="$PROJECT_DIR/frontend"
TAGS="dashboard,production"

cd "$PROJECT_DIR"

# --- Paso 1: Build frontend ---
echo "[1/4] Building frontend..."
(cd "$FRONTEND_DIR" && npm run build --silent)

# --- Paso 2: Build Go binary ---
# CGO_LDFLAGS needed for Wails' macOS production build (UTType framework).
echo "[2/4] Compiling Go binary..."
CGO_LDFLAGS="-framework UniformTypeIdentifiers" go build -tags "$TAGS" -o "$BUILD_DIR/$BINARY_NAME" ./cmd/anvil/
chmod +x "$BUILD_DIR/$BINARY_NAME"

# --- Paso 3: Package .app bundle ---
echo "[3/4] Packaging .app bundle..."
mkdir -p "$MACOS_DIR" "$APP_BUNDLE/Contents/Resources"

# Copy binary into .app
cp "$BUILD_DIR/$BINARY_NAME" "$MACOS_DIR/$BINARY_NAME"
chmod +x "$MACOS_DIR/$BINARY_NAME"

# Info.plist (create if missing)
if [ ! -f "$APP_BUNDLE/Contents/Info.plist" ]; then
  cat > "$APP_BUNDLE/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>$BINARY_NAME</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleIdentifier</key>
    <string>com.ernesto2108.anvil-dashboard</string>
    <key>CFBundleVersion</key>
    <string>1.0.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
PLIST
fi

# --- Paso 4: Codesign ---
echo "[4/4] Signing app..."
codesign --force --deep --sign - "$APP_BUNDLE" 2>/dev/null || true

echo ""
echo "Build complete: $APP_BUNDLE"
echo "Run with:  open \"$APP_BUNDLE\""
echo "Or direct: \"$MACOS_DIR/$BINARY_NAME\" dashboard"
