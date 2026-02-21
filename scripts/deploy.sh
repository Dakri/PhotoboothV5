#!/bin/bash
set -e
TARGET=$1

if [ -z "$TARGET" ]; then
    echo "Usage: ./deploy.sh user@raspberrypi-ip"
    exit 1
fi

# --- SSH ControlMaster: single password entry for all connections ---
CTRL_SOCKET="/tmp/deploy-ssh-$$"
SSH_OPTS="-o StrictHostKeyChecking=no -o ControlMaster=auto -o ControlPath=$CTRL_SOCKET -o ControlPersist=120"

# Clean up control socket on exit
cleanup() {
    ssh -o ControlPath="$CTRL_SOCKET" -O exit "$TARGET" 2>/dev/null || true
}
trap cleanup EXIT

# Open the master connection (this is the ONLY time the password is asked)
echo "🔑 Establishing SSH connection to $TARGET (enter password once)..."
ssh $SSH_OPTS -o ControlMaster=yes -fN "$TARGET"
echo "✅ Connection established!"

echo "🚀 Deploying to $TARGET..."

# 1. Build
./scripts/build-pi.sh

# 1.5 remove old files on target . keep user.conf.json and data
ssh $SSH_OPTS "$TARGET" "sudo rm -rf /opt/photobooth/public && sudo rm -rf /opt/photobooth/scripts && sudo rm -rf /opt/photobooth/photobooth"

# 2. Transfer
echo "📡 Transferring files..."

# Create directory
ssh $SSH_OPTS "$TARGET" "sudo mkdir -p /opt/photobooth && sudo chown -R \$USER:\$USER /opt/photobooth"

# Rsync (uses the same control socket)
rsync -avz --exclude 'data' -e "ssh $SSH_OPTS" dist/ "$TARGET:/opt/photobooth/"

# 3. Setup Service (using -t for sudo in install.sh)
echo "🔧 Setting up service..."
ssh $SSH_OPTS -t "$TARGET" "chmod +x /opt/photobooth/scripts/install.sh && sudo /opt/photobooth/scripts/install.sh"

echo "✅ Deployment Successful!"
