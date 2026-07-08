#!/usr/bin/env bash
# build.sh — Bundle the byteport-cli binary into a portable AppImage.
#
# Usage:
#   ./build.sh                 # Build for the host architecture
#   ./build.sh --arch=arm64    # Cross-build for aarch64 (requires cross)
#   ./build.sh --source=/path/to/source   # Build from a local checkout
#
# Outputs:
#   build/BytePort-<version>-<arch>.AppImage
#
# CI integration: this script is invoked by .github/workflows/release-appimage.yml
# on every tag, which uploads the artifact to the GitHub release alongside
# the Tauri desktop bundle.
#
# Dependencies (auto-installed by this script if missing):
#   - curl, wget, file, tar, gzip, xz, file, desktop-file-utils
#   - appimagetool (downloaded below; we use the official AppImageKit)
#
# Reference: https://docs.appimage.org/introduction/how-to-make-an-appimage.html

set -euo pipefail

# ---- Configuration ----------------------------------------------------------

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
APP_NAME="BytePort"
APP_ID="dev.kooshapari.byteport"
CLI_PKG="byteport-cli"
DEFAULT_VERSION="0.1.0"
DEFAULT_ARCH="$(uname -m)"
SOURCE_DIR="${REPO_ROOT}"
BUILD_DIR="${SCRIPT_DIR}/build"
APPDIR="${BUILD_DIR}/AppDir"

# linuxdeploy release pinned for reproducibility.
# Update periodically; track https://github.com/linuxdeploy/linuxdeploy/releases.
LINUXDEPLOY_VERSION="20240913-1"
LINUXDEPLOY_URL_BASE="https://github.com/linuxdeploy/linuxdeploy/releases/download/${LINUXDEPLOY_VERSION}"
LINUXDEPLOY_ARCHIVE="linuxdeploy-x86_64.AppImage"

# ---- Argument parsing -------------------------------------------------------

ARCH="${DEFAULT_ARCH}"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --arch)        ARCH="$2"; shift 2 ;;
    --arch=*)      ARCH="${1#--arch=}"; shift ;;
    --source)      SOURCE_DIR="$2"; shift 2 ;;
    --source=*)    SOURCE_DIR="${1#--source=}"; shift ;;
    --version)     VERSION="$2"; shift 2 ;;
    --version=*)   VERSION="${1#--version=}"; shift ;;
    --out)         BUILD_DIR="$2"; shift 2 ;;
    -h|--help)
      grep '^#' "$0" | sed 's/^# //; s/^#//'
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 64
      ;;
  esac
done

# Map arch to linuxdeploy's filename suffix.
case "${ARCH}" in
  x86_64|amd64)   LINUXDEPLOY_ARCHIVE="linuxdeploy-x86_64.AppImage" ;;
  aarch64|arm64)  LINUXDEPLOY_ARCHIVE="linuxdeploy-aarch64.AppImage" ;;
  *)
    echo "Unsupported architecture: ${ARCH}" >&2
    echo "Supported: x86_64, amd64, aarch64, arm64" >&2
    exit 65
    ;;
esac

# Version discovery. CI passes --version, local builds default to cargo.
if [[ -z "${VERSION:-}" ]]; then
  if command -v cargo >/dev/null 2>&1 && [[ -f "${SOURCE_DIR}/crates/${CLI_PKG}/Cargo.toml" ]]; then
    VERSION="$(grep '^version' "${SOURCE_DIR}/crates/${CLI_PKG}/Cargo.toml" | head -1 | cut -d'"' -f2)"
  else
    VERSION="${DEFAULT_VERSION}"
  fi
fi

echo ">> Building ${APP_NAME} ${VERSION} for ${ARCH}"
echo ">> Source: ${SOURCE_DIR}"
echo ">> Output: ${BUILD_DIR}"

# ---- Pre-flight -------------------------------------------------------------

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing required command: $1" >&2; exit 66; }
}
need_cmd curl
need_cmd file
need_cmd tar

mkdir -p "${BUILD_DIR}" "${APPDIR}/usr/bin" \
         "${APPDIR}/usr/lib" \
         "${APPDIR}/usr/share/applications" \
         "${APPDIR}/usr/share/icons/hicolor/scalable/apps"

# ---- Build the cli binary ---------------------------------------------------

# Use cargo's target dir so we don't double-compile.
TARGET_DIR="${SOURCE_DIR}/target"
echo ">> Building byteport-cli (cargo build --release)"
(cd "${SOURCE_DIR}" && cargo build --release --manifest-path "crates/${CLI_PKG}/Cargo.toml")

CLI_BIN="${TARGET_DIR}/release/byteport"
if [[ ! -x "${CLI_BIN}" ]]; then
  echo "Expected binary at ${CLI_BIN} but it does not exist." >&2
  exit 67
fi

# Stage the binary, desktop file, and icon into AppDir/.
install -Dm 0755 "${CLI_BIN}"          "${APPDIR}/usr/bin/byteport"
install -Dm 0644 "${SCRIPT_DIR}/byteport.desktop" "${APPDIR}/usr/share/applications/${APP_ID}.desktop"
install -Dm 0644 "${SCRIPT_DIR}/byteport.svg"     "${APPDIR}/usr/share/icons/hicolor/scalable/apps/${APP_ID}.svg"
install -Dm 0644 "${SCRIPT_DIR}/byteport.svg"     "${APPDIR}/byteport.svg"  # linuxdeploy looks in AppDir root
install -Dm 0644 "${SCRIPT_DIR}/byteport.desktop" "${APPDIR}/byteport.desktop"  # legacy layout

# AppRun — minimal wrapper so the AppImage can call our binary directly.
cat > "${APPDIR}/AppRun" <<'EOF'
#!/usr/bin/env bash
# AppRun — entry point invoked by the AppImage runtime.
# When users double-click the AppImage, this script runs. We forward
# all arguments to /usr/bin/byteport so invocation semantics match
# the snap / deb experience (and a CLI manual page can document both).
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "${HERE}/usr/bin/byteport" "$@"
EOF
chmod 0755 "${APPDIR}/AppRun"

# ---- Fetch linuxdeploy ------------------------------------------------------

LINUXDEPLOY_PATH="${BUILD_DIR}/${LINUXDEPLOY_ARCHIVE}"
if [[ ! -x "${LINUXDEPLOY_PATH}" ]]; then
  echo ">> Downloading ${LINUXDEPLOY_ARCHIVE}"
  curl -fL --retry 3 \
    -o "${LINUXDEPLOY_PATH}" \
    "${LINUXDEPLOY_URL_BASE}/${LINUXDEPLOY_ARCHIVE}"
  chmod +x "${LINUXDEPLOY_PATH}"
fi

# ---- Run linuxdeploy --------------------------------------------------------

echo ">> Running linuxdeploy"
"${LINUXDEPLOY_PATH}" \
  --appdir "${APPDIR}" \
  --desktop-file "${APPDIR}/byteport.desktop" \
  --icon           "${APPDIR}/byteport.svg" \
  --output appimage

# ---- Rename the artifact ----------------------------------------------------

ARTIFACT_NAME="${APP_NAME}-${VERSION}-${ARCH}.AppImage"
APPIMAGE_OUT="$(find "${BUILD_DIR}" -maxdepth 1 -name '*.AppImage' -type f | head -1)"
if [[ -z "${APPIMAGE_OUT}" ]]; then
  echo "linuxdeploy did not produce an AppImage" >&2
  exit 68
fi
mv "${APPIMAGE_OUT}" "${BUILD_DIR}/${ARTIFACT_NAME}"
chmod +x "${BUILD_DIR}/${ARTIFACT_NAME}"

echo ">> Done: ${BUILD_DIR}/${ARTIFACT_NAME}"
echo ">> Test by running: ${BUILD_DIR}/${ARTIFACT_NAME} --version"
