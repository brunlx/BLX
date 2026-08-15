#!/usr/bin/env bash
set -euo pipefail

# Builds an installable APK of BLX without Gradle, using only the Android SDK
# command-line tools (aapt2, d8, zipalign, apksigner) + JDK keytool.
#
# Usage:   VERSION=0.2.3 ./build-apk.sh
# Env:     ANDROID_HOME  (defaults to $HOME/Android/Sdk)
#          VERSION       (defaults to latest git tag, v stripped)
# Output:  BLX-<version>.apk in the android/ directory.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

ANDROID_HOME="${ANDROID_HOME:-$HOME/Android/Sdk}"
SDKMANAGER="$ANDROID_HOME/cmdline-tools/latest/bin/sdkmanager"
BUILD_TOOLS="$ANDROID_HOME/build-tools/36.0.0"
PLATFORM="$ANDROID_HOME/platforms/android-34"
BT="$BUILD_TOOLS"

VERSION="${VERSION:-$(git -C "$REPO_DIR" describe --tags --abbrev=0 2>/dev/null || echo dev)}"
VERSION="${VERSION#v}"
VERSION_CODE="${VERSION_CODE:-1}"
KEYSTORE="$SCRIPT_DIR/keystore.jks"
KEYSTORE_PASS="${KEYSTORE_PASS:-blx-apk-$(sha1sum /dev/null | cut -c1-4)}"
APK_OUT="$SCRIPT_DIR/BLX-$VERSION.apk"

ensure_sdk() {
  if [ -x "$BT/aapt2" ] && [ -f "$PLATFORM/android.jar" ]; then
    return
  fi
  echo "==> Instalando Android SDK em $ANDROID_HOME (primeira execução)..."
  mkdir -p "$ANDROID_HOME/cmdline-tools"
  TMP_ZIP="$(mktemp --suffix=.zip)"
  curl -fsSL -o "$TMP_ZIP" \
    https://dl.google.com/android/repository/commandlinetools-linux-11076708_latest.zip
  unzip -qo "$TMP_ZIP" -d "$ANDROID_HOME/cmdline-tools"
  mv "$ANDROID_HOME/cmdline-tools/cmdline-tools" "$ANDROID_HOME/cmdline-tools/latest"
  rm -f "$TMP_ZIP"
  yes | "$SDKMANAGER" --licenses >/dev/null 2>&1 || true
  "$SDKMANAGER" "platform-tools" "platforms;android-34" "build-tools;36.0.0" >/dev/null
}

build_go_binary() {
  echo "==> Compilando servidor BLX (GOOS=android GOARCH=arm64)..."
  (cd "$REPO_DIR" && \
   GOOS=android GOARCH=arm64 CGO_ENABLED=0 \
   go build -trimpath -ldflags "-s -w" \
   -o "$SCRIPT_DIR/stage/assets/blx-server" ./cmd/server)
}

ensure_keystore() {
  if [ -f "$KEYSTORE" ]; then
    return
  fi
  echo "==> Gerando keystore de assinatura..."
  keytool -genkeypair -keystore "$KEYSTORE" -storepass "$KEYSTORE_PASS" \
    -keyalg RSA -keysize 2048 -validity 10950 -alias blx \
    -dname "CN=BLX, OU=BLX, O=BLX, C=BR" -keypass "$KEYSTORE_PASS" >/dev/null 2>&1
}

build() {
  ensure_sdk
  ensure_keystore
  mkdir -p "$SCRIPT_DIR/stage/assets" "$SCRIPT_DIR/stage/classes" "$SCRIPT_DIR/stage/dex"
  build_go_binary

  echo "==> Compilando MainActivity..."
  javac --release 8 -encoding UTF-8 \
    -classpath "$PLATFORM/android.jar" \
    -d "$SCRIPT_DIR/stage/classes" \
    "$SCRIPT_DIR/app/src/main/java/com/brunlx/blx/MainActivity.java"

  echo "==> Gerando classes.dex..."
  "$BT/d8" --release --lib "$PLATFORM/android.jar" \
    --output "$SCRIPT_DIR/stage/dex" \
    $(find "$SCRIPT_DIR/stage/classes" -name '*.class')

  echo "==> Empacotando recursos e manifesto..."
  "$BT/aapt2" link -o "$SCRIPT_DIR/stage/unsigned.apk" \
    --manifest "$SCRIPT_DIR/app/src/main/AndroidManifest.xml" \
    -I "$PLATFORM/android.jar" \
    --min-sdk-version 24 --target-sdk-version 34 \
    --version-code "$VERSION_CODE" --version-name "$VERSION"

  echo "==> Adicionando classes.dex (stored) e assets..."
  cp "$SCRIPT_DIR/stage/unsigned.apk" "$SCRIPT_DIR/stage/with-assets.apk"
  python3 - "$SCRIPT_DIR/stage" <<'PY'
import os
import sys
import zipfile

stage = sys.argv[1]
apk = os.path.join(stage, "with-assets.apk")
with zipfile.ZipFile(apk, "a", allowZip64=True) as zf:
    zf.write(os.path.join(stage, "dex", "classes.dex"), "classes.dex",
             compress_type=zipfile.ZIP_STORED)
    for root, _dirs, files in os.walk(os.path.join(stage, "assets")):
        for f in files:
            full = os.path.join(root, f)
            arc = os.path.relpath(full, stage)
            zf.write(full, arc, compress_type=zipfile.ZIP_DEFLATED)
PY

  echo "==> Alinhando (zipalign)..."
  "$BT/zipalign" -f 4 "$SCRIPT_DIR/stage/with-assets.apk" "$SCRIPT_DIR/stage/aligned.apk"

  echo "==> Assinando..."
  "$BT/apksigner" sign --ks "$KEYSTORE" --ks-pass "pass:$KEYSTORE_PASS" \
    --ks-key-alias blx --key-pass "pass:$KEYSTORE_PASS" \
    --out "$APK_OUT" "$SCRIPT_DIR/stage/aligned.apk"

  echo "==> Verificando..."
  "$BT/apksigner" verify --print-certs "$APK_OUT"

  rm -rf "$SCRIPT_DIR/stage"
  echo
  echo "APK gerado: $APK_OUT"
  ls -lh "$APK_OUT"
}

build
