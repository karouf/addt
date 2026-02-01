#!/bin/bash

echo "=== Firewall Test ==="
echo ""

echo "1. Testing allowed domain (api.anthropic.com)..."
if curl --max-time 5 -s https://api.anthropic.com > /dev/null 2>&1; then
    echo "✓ api.anthropic.com - ALLOWED (expected)"
else
    echo "✗ api.anthropic.com - BLOCKED (unexpected!)"
fi

echo ""
echo "2. Testing blocked domain (google.com)..."
if curl --max-time 5 -s https://google.com > /dev/null 2>&1; then
    echo "✗ google.com - ALLOWED (unexpected!)"
else
    echo "✓ google.com - BLOCKED (expected)"
fi

echo ""
echo "3. Testing another blocked domain (example.com)..."
if curl --max-time 5 -s https://example.com > /dev/null 2>&1; then
    echo "✗ example.com - ALLOWED (unexpected!)"
else
    echo "✓ example.com - BLOCKED (expected)"
fi

echo ""
echo "4. Checking iptables rules..."
sudo iptables -L OUTPUT -n -v | grep -E "Chain OUTPUT|allowed_ips|LOG|DROP"

echo ""
echo "5. Checking ipset..."
sudo ipset list allowed_ips | head -20
