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
KEYSTORE_PASS_FILE="$SCRIPT_DIR/.keystore-pass"
APK_OUT="$SCRIPT_DIR/BLX-$VERSION.apk"

# A senha do keystore de assinatura é aleatória (nunca determinística), gerada
# na primeira execução e persistida em .keystore-pass (gitignored, chmod 600).
# Para CI/reprodução, sobrescreva com KEYSTORE_PASS e gere o keystore uma vez.
resolve_keystore_pass() {
  if [ -n "${KEYSTORE_PASS:-}" ]; then
    return
  fi
  if [ -f "$KEYSTORE_PASS_FILE" ]; then
    KEYSTORE_PASS="$(cat "$KEYSTORE_PASS_FILE")"
    return
  fi
  if [ -f "$KEYSTORE" ]; then
    echo "!! Keystore existente sem senha salva em $KEYSTORE_PASS_FILE." >&2
    echo "   Defina a variável KEYSTORE_PASS com a senha usada ao gerá-lo." >&2
    exit 1
  fi
  KEYSTORE_PASS="$(LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 24)"
}

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
  echo "==> Compilando servidor BLX (GOOS=android, multi-ABI)..."
  # O binário é empacotado como lib nativa (jniLibs/<abi>/libblxserver.so): o
  # instalador extrai para nativeLibraryDir, único lugar executável desde o
  # Android 10 (SELinux W^X bloqueia exec em getFilesDir()).
  # GOOS=android sem cgo (NDK) suporta apenas arm64; arm/amd64 exigiriam
  # compilador cruzado. arm64-v8a cobre a esmagadora maioria dos dispositivos.
  build_abi arm64-v8a arm64 ""
}

build_abi() {
  local abi="$1" arch="$2" goarm="$3"
  local out="$SCRIPT_DIR/stage/jniLibs/$abi/libblxserver.so"
  mkdir -p "$(dirname "$out")"
  (cd "$REPO_DIR" && \
   GOOS=android GOARCH="$arch" GOARM="$goarm" CGO_ENABLED=0 \
   go build -trimpath -ldflags "-s -w" \
   -o "$out" ./cmd/server)
  echo "   $abi: $(du -h "$out" | cut -f1)"
}

ensure_keystore() {
  if [ -f "$KEYSTORE" ]; then
    return
  fi
  echo "==> Gerando keystore de assinatura..."
  keytool -genkeypair -keystore "$KEYSTORE" -storepass "$KEYSTORE_PASS" \
    -keyalg RSA -keysize 2048 -validity 10950 -alias blx \
    -dname "CN=BLX, OU=BLX, O=BLX, C=BR" -keypass "$KEYSTORE_PASS" >/dev/null 2>&1
  umask 077
  printf '%s' "$KEYSTORE_PASS" > "$KEYSTORE_PASS_FILE"
}

build() {
  resolve_keystore_pass
  ensure_sdk
  ensure_keystore
  mkdir -p "$SCRIPT_DIR/stage/classes" "$SCRIPT_DIR/stage/dex" "$SCRIPT_DIR/stage/res"
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

  echo "==> Compilando resources..."
  "$BT/aapt2" compile --dir "$SCRIPT_DIR/app/src/main/res" \
    -o "$SCRIPT_DIR/stage/res" >/dev/null

  echo "==> Empacotando recursos e manifesto..."
  "$BT/aapt2" link -o "$SCRIPT_DIR/stage/unsigned.apk" \
    --manifest "$SCRIPT_DIR/app/src/main/AndroidManifest.xml" \
    -I "$PLATFORM/android.jar" \
    --min-sdk-version 24 --target-sdk-version 34 \
    --version-code "$VERSION_CODE" --version-name "$VERSION" \
    $(find "$SCRIPT_DIR/stage/res" -name '*.flat')

  echo "==> Adicionando classes.dex (stored) e lib nativa (stored)..."
  cp "$SCRIPT_DIR/stage/unsigned.apk" "$SCRIPT_DIR/stage/with-libs.apk"
  python3 - "$SCRIPT_DIR/stage" <<'PY'
import os
import sys
import zipfile

stage = sys.argv[1]
apk = os.path.join(stage, "with-libs.apk")
with zipfile.ZipFile(apk, "a", allowZip64=True) as zf:
    zf.write(os.path.join(stage, "dex", "classes.dex"), "classes.dex",
             compress_type=zipfile.ZIP_STORED)
    for abi in sorted(os.listdir(os.path.join(stage, "jniLibs"))):
        for f in os.listdir(os.path.join(stage, "jniLibs", abi)):
            full = os.path.join(stage, "jniLibs", abi, f)
            zf.write(full, "lib/%s/%s" % (abi, f), compress_type=zipfile.ZIP_STORED)
PY

  echo "==> Alinhando (zipalign)..."
  "$BT/zipalign" -f 4 "$SCRIPT_DIR/stage/with-libs.apk" "$SCRIPT_DIR/stage/aligned.apk"

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
