#!/bin/bash
set -e

TARGET=$1

if [ -z "$TARGET" ]; then
    echo "Usage: ./deploy.sh user@raspberrypi-ip"
    exit 1
fi

echo "🚀 Deploying to $TARGET..."

# 1. Build
./scripts/build-pi.sh

# 2. Transfer
echo "📡 Transferring files..."
# Create directory if not exists
ssh $TARGET "sudo mkdir -p /opt/photobooth && sudo chown -R $USER:$USER /opt/photobooth"

# Rsync (exclude data)
rsync -avz --exclude 'data' dist/ $TARGET:/opt/photobooth/

# 3. Setup Service (First time or update)
echo "🔧 Setting up service..."
ssh $TARGET "chmod +x /opt/photobooth/scripts/install.sh && sudo /opt/photobooth/scripts/install.sh"

echo "✅ Deployment Successful!"
