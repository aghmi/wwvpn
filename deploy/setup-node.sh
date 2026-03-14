#!/bin/bash
set -euo pipefail

NODE_IP=$(curl -s ifconfig.me)
AWG_PORT=51820
XRAY_PORT=443
GRPC_PORT=50051
AWG_INTERFACE="awg0"
AWG_SUBNET="10.0.0.0/24"
AWG_SERVER_IP="10.0.0.1"
XRAY_SNI="www.microsoft.com"

SERVER_NIC=$(ip -4 route ls | grep default | grep -Po '(?<=dev )(\S+)' | head -1)

echo "=== wwvpn Node Setup ==="
echo "Server IP: $NODE_IP"
echo "Network interface: $SERVER_NIC"

apt-get update
apt-get install -y software-properties-common curl wget unzip iptables linux-headers-$(uname -r) gnupg2

echo "=== Enabling deb-src repos ==="
if [[ -e /etc/apt/sources.list.d/ubuntu.sources ]]; then
    if ! grep -q "Types: deb-src" /etc/apt/sources.list.d/ubuntu.sources; then
        cp /etc/apt/sources.list.d/ubuntu.sources /etc/apt/sources.list.d/amneziawg.sources
        sed -i 's/Types: deb/Types: deb-src/' /etc/apt/sources.list.d/amneziawg.sources
    fi
elif [[ -e /etc/apt/sources.list ]]; then
    if ! grep -q "^deb-src" /etc/apt/sources.list; then
        cp /etc/apt/sources.list /etc/apt/sources.list.d/amneziawg.sources.list
        sed -i 's/^deb/deb-src/' /etc/apt/sources.list.d/amneziawg.sources.list
    fi
fi

echo "=== Installing AmneziaWG ==="
if add-apt-repository -y ppa:amnezia/ppa; then
    apt-get update
    apt-get install -y amneziawg amneziawg-tools
else
    echo "PPA failed, building from source..."
    apt-get install -y git dkms build-essential
    cd /tmp
    git clone https://github.com/amnezia-vpn/amneziawg-linux-kernel-module.git
    cd amneziawg-linux-kernel-module/src
    make && make install
    cd /tmp
    git clone https://github.com/amnezia-vpn/amneziawg-tools.git
    cd amneziawg-tools/src
    make && make install
    cd /
    depmod -a
    modprobe amneziawg
fi

mkdir -p /etc/amnezia/amneziawg

AWG_PRIVATE_KEY=$(awg genkey)
AWG_PUBLIC_KEY=$(echo "$AWG_PRIVATE_KEY" | awg pubkey)

JC=$(shuf -i3-10 -n1)
JMIN=$(shuf -i50-100 -n1)
JMAX=$(shuf -i500-1000 -n1)

gen_s() {
    local s1 s2
    while true; do
        s1=$(shuf -i15-150 -n1)
        s2=$(shuf -i15-150 -n1)
        if (( s1 + 56 != s2 )); then
            echo "$s1 $s2"
            return
        fi
    done
}
read S1 S2 <<< $(gen_s)

gen_h() {
    local h1 h2 h3 h4
    while true; do
        h1=$(shuf -i5-2147483647 -n1)
        h2=$(shuf -i5-2147483647 -n1)
        h3=$(shuf -i5-2147483647 -n1)
        h4=$(shuf -i5-2147483647 -n1)
        if (( h1 != h2 && h1 != h3 && h1 != h4 && h2 != h3 && h2 != h4 && h3 != h4 )); then
            echo "$h1 $h2 $h3 $h4"
            return
        fi
    done
}
read H1 H2 H3 H4 <<< $(gen_h)

cat > /etc/amnezia/amneziawg/$AWG_INTERFACE.conf << AWGEOF
[Interface]
PrivateKey = $AWG_PRIVATE_KEY
Address = $AWG_SERVER_IP/24
ListenPort = $AWG_PORT
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o $SERVER_NIC -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o $SERVER_NIC -j MASQUERADE
Jc = $JC
Jmin = $JMIN
Jmax = $JMAX
S1 = $S1
S2 = $S2
H1 = $H1
H2 = $H2
H3 = $H3
H4 = $H4
AWGEOF

chmod 600 /etc/amnezia/amneziawg/$AWG_INTERFACE.conf

echo "net.ipv4.ip_forward = 1" > /etc/sysctl.d/99-wwvpn.conf
echo "net.ipv6.conf.all.forwarding = 1" >> /etc/sysctl.d/99-wwvpn.conf
sysctl -p /etc/sysctl.d/99-wwvpn.conf

systemctl enable awg-quick@$AWG_INTERFACE
systemctl start awg-quick@$AWG_INTERFACE

echo "=== Installing XRay ==="
bash <(curl -Ls https://raw.githubusercontent.com/XTLS/Xray-install/main/install-release.sh)

XRAY_KEYS=$(xray x25519)
XRAY_PRIVATE_KEY=$(echo "$XRAY_KEYS" | grep "Private" | awk '{print $3}')
XRAY_PUBLIC_KEY=$(echo "$XRAY_KEYS" | grep "Public" | awk '{print $3}')
XRAY_SHORT_ID=$(openssl rand -hex 8)

cat > /usr/local/etc/xray/config.json << XRAYEOF
{
  "log": {
    "loglevel": "warning"
  },
  "api": {
    "tag": "api",
    "services": ["HandlerService"]
  },
  "inbounds": [
    {
      "tag": "api-in",
      "listen": "127.0.0.1",
      "port": 10085,
      "protocol": "dokodemo-door",
      "settings": {
        "address": "127.0.0.1"
      }
    },
    {
      "tag": "vless-in",
      "listen": "0.0.0.0",
      "port": $XRAY_PORT,
      "protocol": "vless",
      "settings": {
        "clients": [],
        "decryption": "none"
      },
      "streamSettings": {
        "network": "tcp",
        "security": "reality",
        "realitySettings": {
          "dest": "$XRAY_SNI:443",
          "serverNames": ["$XRAY_SNI"],
          "privateKey": "$XRAY_PRIVATE_KEY",
          "shortIds": ["$XRAY_SHORT_ID"]
        }
      }
    }
  ],
  "outbounds": [
    {
      "protocol": "freedom",
      "tag": "direct"
    }
  ],
  "routing": {
    "rules": [
      {
        "inboundTag": ["api-in"],
        "outboundTag": "api",
        "type": "field"
      }
    ]
  }
}
XRAYEOF

systemctl enable xray
systemctl restart xray

echo "=== Configuring Firewall ==="
ufw allow $AWG_PORT/udp
ufw allow $XRAY_PORT/tcp
ufw allow $GRPC_PORT/tcp
ufw allow 22/tcp
ufw --force enable

echo "=== BBR Congestion Control ==="
echo "net.core.default_qdisc=fq" >> /etc/sysctl.d/99-wwvpn.conf
echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.d/99-wwvpn.conf
sysctl -p /etc/sysctl.d/99-wwvpn.conf

mkdir -p /opt/wwvpn

echo ""
echo "============================================"
echo "=== wwvpn Node Setup Complete ==="
echo "============================================"
echo ""
echo "SERVER_IP=$NODE_IP"
echo "AWG_PORT=$AWG_PORT"
echo "AWG_PUBLIC_KEY=$AWG_PUBLIC_KEY"
echo "AWG_JC=$JC"
echo "AWG_JMIN=$JMIN"
echo "AWG_JMAX=$JMAX"
echo "AWG_S1=$S1"
echo "AWG_S2=$S2"
echo "AWG_H1=$H1"
echo "AWG_H2=$H2"
echo "AWG_H3=$H3"
echo "AWG_H4=$H4"
echo "XRAY_PORT=$XRAY_PORT"
echo "XRAY_PUBLIC_KEY=$XRAY_PUBLIC_KEY"
echo "XRAY_SHORT_ID=$XRAY_SHORT_ID"
echo "XRAY_SNI=$XRAY_SNI"
echo "GRPC_PORT=$GRPC_PORT"
echo ""
echo "Save these values! You need them for the API config."
echo ""

cat > /opt/wwvpn/node-config.env << ENVEOF
SERVER_IP=$NODE_IP
AWG_PORT=$AWG_PORT
AWG_PUBLIC_KEY=$AWG_PUBLIC_KEY
AWG_PRIVATE_KEY=$AWG_PRIVATE_KEY
AWG_JC=$JC
AWG_JMIN=$JMIN
AWG_JMAX=$JMAX
AWG_S1=$S1
AWG_S2=$S2
AWG_H1=$H1
AWG_H2=$H2
AWG_H3=$H3
AWG_H4=$H4
XRAY_PORT=$XRAY_PORT
XRAY_PUBLIC_KEY=$XRAY_PUBLIC_KEY
XRAY_PRIVATE_KEY=$XRAY_PRIVATE_KEY
XRAY_SHORT_ID=$XRAY_SHORT_ID
XRAY_SNI=$XRAY_SNI
GRPC_PORT=$GRPC_PORT
ENVEOF

chmod 600 /opt/wwvpn/node-config.env
echo "Config saved to /opt/wwvpn/node-config.env"
