#!/bin/bash
# ============================================================
# AURA Core — VPS Deploy Script (IDCloudHost / Ubuntu)
# Jalankan sekali di VPS baru: bash deploy-vps.sh
# ============================================================
set -e

echo "╔══════════════════════════════════════╗"
echo "║   AURA Core VPS Deployment Script    ║"
echo "╚══════════════════════════════════════╝"

# 1. Install Go
echo "[1/5] Installing Go..."
if ! command -v go &>/dev/null; then
    wget -q https://go.dev/dl/go1.23.1.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz
    rm go1.23.1.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
    export PATH=$PATH:/usr/local/go/bin
fi
go version

# 2. Clone repo
echo "[2/5] Cloning repository..."
cd ~
if [ -d "IMUII-Go" ]; then
    cd IMUII-Go && git pull
else
    git clone https://github.com/848ill/IMUII-Go.git
    cd IMUII-Go
fi

# 3. Setup env
echo "[3/5] Setting up environment..."
if [ ! -f aura-core/.env ]; then
    cp aura-core/.env.example aura-core/.env
    echo ""
    echo "⚠️  EDIT .env DENGAN API KEYS KAMU:"
    echo "    nano ~/IMUII-Go/aura-core/.env"
    echo ""
    echo "Setelah edit, jalankan ulang script ini."
    exit 0
fi

# 4. Build
echo "[4/5] Building AURA Core..."
cd aura-core
go build -ldflags="-s -w" -o bin/aura-core ./cmd/server
echo "Binary size: $(du -h bin/aura-core | cut -f1)"

# 5. Setup systemd service (auto-restart, auto-start on boot)
echo "[5/5] Setting up systemd service..."
sudo tee /etc/systemd/system/aura-core.service > /dev/null <<EOF
[Unit]
Description=AURA Core - Academic RAG Assistant
After=network.target

[Service]
Type=simple
User=$USER
WorkingDir=$HOME/IMUII-Go/aura-core
ExecStart=$HOME/IMUII-Go/aura-core/bin/aura-core
Restart=always
RestartSec=5
EnvironmentFile=$HOME/IMUII-Go/aura-core/.env

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable aura-core
sudo systemctl restart aura-core

echo ""
echo "╔══════════════════════════════════════╗"
echo "║       ✅ DEPLOY SUKSES!              ║"
echo "║  AURA Core running on port 8090     ║"
echo "║  Status: sudo systemctl status aura-core"
echo "║  Logs:   sudo journalctl -u aura-core -f"
echo "╚══════════════════════════════════════╝"
