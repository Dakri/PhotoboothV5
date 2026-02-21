#!/bin/bash
set -e

echo "🏗️  Building Photobooth V5 for Raspberry Pi (ARM64)..."

# 0. Clean workspace
echo "🗑️  Removing old dist directory..."
rm -rf dist

# 1. Build Frontend
echo "📦 Building Frontend... "
cd frontend
npm run build
cd ..

# 2. Build Backend (Cross-Compile)
echo "🐹 Building Backend (Go)..."
cd backend
go mod tidy
# Enable CGO? No, we want static binary if possible, but imaging might need it? 
# pure go imaging lib used, so CGO_ENABLED=0 should work.
# However, if we ever needed sqlite or similar, we'd need CGO.
# For now, pure Go.
env GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ../dist/photobooth ./cmd/server
cd ..


# 3. Copy Static Assets
echo "📂 Copying assets..."
mkdir -p dist/public/frontend
mkdir -p dist/public/legacy
mkdir -p dist/config

# Frontend built to ../public/frontend (via vite config we set earlier? Let's check)
# Vite config said: outDir: '../public/frontend' relative to frontend dir.
# So it should be in public/frontend already? No, vite config root is frontend/.
# So '../public/frontend' means 'photobooth/public/frontend'.
# Backend Main.go expects './public/frontend'.
# So dist structure should be:
# dist/
#   photobooth (binary)
#   public/
#     frontend/
#     legacy/
#   config/
#   scripts/

# Copy Frontend (already in public/frontend if vite config worked, but let's ensure)
# Actually, let's just copy from wherever vite put it. 
# Our Vite config was: outDir: '../public/frontend'.
# So running build in frontend/ puts files in photobooth/public/frontend.
# We want to move everything to dist/ for the release.


cp -r public dist/
cp -r legacy dist/public/
cp config/server.conf.json dist/config/
cp -r scripts dist/

echo "✅ Build Complete! artifacts are in dist/"
