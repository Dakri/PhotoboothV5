#!/bin/bash
set -e

echo "🔧 Installing Photobooth Service..."

# 1. Install Dependencies
echo "📦 Checking system dependencies..."
if ! command -v gphoto2 &> /dev/null; then
    echo "Installing gphoto2..."
    apt-get update && apt-get install -y gphoto2
fi

if ! command -v nmcli &> /dev/null; then
    echo "⚠️ NetworkManager (nmcli) not found. WiFi Hotspot feature will not work!"
fi

# 2. Setup Directories
mkdir -p /opt/photobooth/data/photos/original
mkdir -p /opt/photobooth/data/photos/preview
mkdir -p /opt/photobooth/data/photos/thumb
chmod -R 777 /opt/photobooth/data # Allow storage access

# 3. Setup Systemd Service
echo "📝 Configuring systemd..."
cp /opt/photobooth/scripts/photobooth.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable photobooth
systemctl restart photobooth

echo "✅ Installation Complete! Service is running."
echo "   Access Dashboard: http://192.168.4.1"
