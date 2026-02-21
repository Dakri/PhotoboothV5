#!/bin/bash
set -e

echo "🔧 Installing Photobooth Service..."

# 1. Install Dependencies
echo "📦 Checking system dependencies..."
if ! command -v gphoto2 &> /dev/null || ! dpkg -s dnsmasq-base &> /dev/null; then
    echo "Installing dependencies (gphoto2, dnsmasq-base)..."
    apt-get update && apt-get install -y gphoto2 dnsmasq-base
fi

# 1a. Install EPEG (fast JPEG processing)
if ! command -v epeg &> /dev/null; then
    echo "⚡ Installing EPEG (fast JPEG scaler)..."
    apt-get update
    apt-get install -y git build-essential autoconf libtool pkg-config libjpeg-dev libexif-dev libpopt-dev

    cd /tmp
    rm -rf epeg
    git clone https://github.com/mattes/epeg.git
    cd epeg
    ./autogen.sh
    ./configure --prefix=/usr
    make
    make install
    ldconfig
    cd ..
    rm -rf epeg
    echo "   EPEG installed successfully."
else
    echo "   EPEG already installed."
fi

if ! command -v nmcli &> /dev/null; then
    echo "⚠️ NetworkManager (nmcli) not found. WiFi Hotspot feature will not work!"
fi

# 1b. Disable gvfs-gphoto2-volume-monitor (blocks gphoto2 USB access)
echo "🔒 Disabling gvfs-gphoto2-volume-monitor..."
pkill -f gvfs-gphoto2 2>/dev/null || true
# Prevent it from starting again
if [ -f /usr/lib/gvfs/gvfs-gphoto2-volume-monitor ]; then
    chmod -x /usr/lib/gvfs/gvfs-gphoto2-volume-monitor
    echo "   Disabled /usr/lib/gvfs/gvfs-gphoto2-volume-monitor"
fi

# 2. Setup Directories
chmod -R 777 /opt/photobooth/data # Allow storage access

# 3. Setup Systemd Service
echo "📝 Configuring systemd..."
cp /opt/photobooth/scripts/photobooth.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable photobooth
systemctl restart photobooth

echo "✅ Installation Complete! Service is running."
echo "   Access Dashboard: http://192.168.4.1"
