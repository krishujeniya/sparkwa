#!/usr/bin/env bash
# SparkWA — Release Packaging Script
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

VERSION="${VERSION:-$(git -C "$PROJECT_ROOT" describe --tags --exact-match 2>/dev/null || echo "v1.0.0")}"
COMMIT=$(git -C "$PROJECT_ROOT" rev-parse --short HEAD 2>/dev/null || echo "release")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

DIST_DIR="$PROJECT_ROOT/dist"
mkdir -p "$DIST_DIR"

LDFLAGS="-s -w -X main.Version=$VERSION -X main.GitCommit=$COMMIT -X main.BuildDate=$BUILD_DATE"

echo "================================================================================"
echo " Packaging SparkWA $VERSION ($COMMIT)"
echo "================================================================================"

for ARCH in amd64 arm64; do
    echo "Compiling for linux/$ARCH..."
    OUTPUT_BIN="$PROJECT_ROOT/bin/linux-$ARCH/sparkwa"
    mkdir -p "$(dirname "$OUTPUT_BIN")"

    CGO_ENABLED=0 GOOS=linux GOARCH=$ARCH go build \
        -trimpath \
        -ldflags="$LDFLAGS" \
        -o "$OUTPUT_BIN" \
        "$PROJECT_ROOT/cmd/sparkwa"

    ARCHIVE_NAME="sparkwa-$VERSION-linux-$ARCH.tar.gz"
    STAGE_DIR="$DIST_DIR/stage-$ARCH"
    rm -rf "$STAGE_DIR"
    mkdir -p "$STAGE_DIR/sparkwa"

    # Copy binary
    cp "$OUTPUT_BIN" "$STAGE_DIR/sparkwa/sparkwa"
    chmod 0755 "$STAGE_DIR/sparkwa/sparkwa"

    # Copy deployment & scripts
    mkdir -p "$STAGE_DIR/sparkwa/deploy" "$STAGE_DIR/sparkwa/scripts" "$STAGE_DIR/sparkwa/docs"
    cp "$PROJECT_ROOT/deploy/sparkwa.service" "$STAGE_DIR/sparkwa/deploy/"
    cp "$PROJECT_ROOT/scripts/"*.sh "$STAGE_DIR/sparkwa/scripts/"
    chmod 0755 "$STAGE_DIR/sparkwa/scripts/"*.sh

    # Copy documentation & configuration
    cp -r "$PROJECT_ROOT/docs/"* "$STAGE_DIR/sparkwa/docs/"
    cp "$PROJECT_ROOT/.env.example" "$STAGE_DIR/sparkwa/"
    cp "$PROJECT_ROOT/README.md" "$STAGE_DIR/sparkwa/"
    cp "$PROJECT_ROOT/LICENSE" "$STAGE_DIR/sparkwa/"
    cp "$PROJECT_ROOT/CHANGELOG.md" "$STAGE_DIR/sparkwa/"

    # Create tar.gz archive
    echo "Creating archive $ARCHIVE_NAME..."
    tar -czf "$DIST_DIR/$ARCHIVE_NAME" -C "$STAGE_DIR" sparkwa
    rm -rf "$STAGE_DIR"

    echo "✓ Built $ARCHIVE_NAME"

    # Build Debian .deb package if dpkg-deb is available
    if command -v dpkg-deb >/dev/null 2>&1; then
        DEB_STAGE="$DIST_DIR/deb-$ARCH"
        rm -rf "$DEB_STAGE"
        mkdir -p "$DEB_STAGE/DEBIAN" "$DEB_STAGE/usr/local/bin" "$DEB_STAGE/etc/sparkwa" "$DEB_STAGE/etc/systemd/system"
        cp "$OUTPUT_BIN" "$DEB_STAGE/usr/local/bin/sparkwa"
        chmod 0755 "$DEB_STAGE/usr/local/bin/sparkwa"
        cp "$PROJECT_ROOT/deploy/sparkwa.service" "$DEB_STAGE/etc/systemd/system/"
        cp "$PROJECT_ROOT/.env.example" "$DEB_STAGE/etc/sparkwa/config.env.example"

        CLEAN_VERSION="${VERSION#v}"
        cat <<EOF > "$DEB_STAGE/DEBIAN/control"
Package: sparkwa
Version: $CLEAN_VERSION
Section: utils
Priority: optional
Architecture: $ARCH
Maintainer: Krish Ujeniya <social.krishujeniya@gmail.com>
Description: Headless WhatsApp to Google Gemini Spark Bridge
 SparkWA connects WhatsApp to Google Gemini Spark through an event-driven
 Gmail trigger schedule, providing autonomous background task execution.
EOF

        cat <<'EOF' > "$DEB_STAGE/DEBIAN/postinst"
#!/bin/sh
set -e
if ! id "sparkwa" >/dev/null 2>&1; then
    useradd --system --no-create-home --user-group --shell /usr/sbin/nologin sparkwa || true
fi
mkdir -p /var/lib/sparkwa /etc/sparkwa
chown -R sparkwa:sparkwa /var/lib/sparkwa /etc/sparkwa
chmod 0700 /var/lib/sparkwa
systemctl daemon-reload || true
exit 0
EOF
        chmod 0755 "$DEB_STAGE/DEBIAN/postinst"

        DEB_NAME="sparkwa_${CLEAN_VERSION}_${ARCH}.deb"
        echo "Building Debian package $DEB_NAME..."
        dpkg-deb --build --root-owner-group "$DEB_STAGE" "$DIST_DIR/$DEB_NAME"
        rm -rf "$DEB_STAGE"
        echo "✓ Built $DEB_NAME"
    fi
done

# Compute checksums
echo "Generating SHA-256 checksums..."
cd "$DIST_DIR"
sha256sum sparkwa-*.tar.gz sparkwa_*.deb > SHA256SUMS

echo "================================================================================"
cat "$DIST_DIR/SHA256SUMS"
echo "================================================================================"
echo "✓ Release packaging complete in $DIST_DIR"
