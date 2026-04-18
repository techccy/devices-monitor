#!/bin/bash

# CCY Agent Service Installation Script

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVICE_TYPE=${1:-systemd}

if [ "$SERVICE_TYPE" != "systemd" ] && [ "$SERVICE_TYPE" != "supervisord" ]; then
    echo "Usage: $0 [systemd|supervisord]"
    echo "Default: systemd"
    exit 1
fi

echo "Installing CCY Agent as $SERVICE_TYPE service..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root"
    exit 1
fi

# Create directories
echo "Creating directories..."
mkdir -p /opt/ccy/bin
mkdir -p /opt/ccy/config
mkdir -p /var/log

# Copy binaries
echo "Copying binaries..."
cp "$PROJECT_ROOT/bin/ccy-agent" /opt/ccy/bin/
chmod +x /opt/ccy/bin/ccy-agent

# Copy example config if not exists
if [ ! -f /opt/ccy/config/agent.json ]; then
    echo "Copying example configuration..."
    cp "$PROJECT_ROOT/config/agent.example.json" /opt/ccy/config/agent.json
    echo "Please edit /opt/ccy/config/agent.json with your device credentials"
fi

if [ "$SERVICE_TYPE" = "systemd" ]; then
    # Install systemd service
    echo "Installing systemd service..."
    cp "$SCRIPT_DIR/ccy-agent.service" /etc/systemd/system/
    systemctl daemon-reload
    systemctl enable ccy-agent.service
    echo "Service installed. Use: systemctl start ccy-agent"
    echo "To check status: systemctl status ccy-agent"
    echo "To view logs: journalctl -u ccy-agent -f"
elif [ "$SERVICE_TYPE" = "supervisord" ]; then
    # Install supervisord config
    echo "Installing supervisord configuration..."
    cp "$SCRIPT_DIR/ccy-agent.conf" /etc/supervisor/conf.d/
    supervisorctl reread
    supervisorctl update
    echo "Service installed. Use: supervisorctl start ccy-agent"
    echo "To check status: supervisorctl status ccy-agent"
    echo "To view logs: supervisorctl tail -f ccy-agent"
fi

echo ""
echo "Installation completed!"
echo "Please edit the configuration file before starting the service:"
echo "  /opt/ccy/config/agent.json"
