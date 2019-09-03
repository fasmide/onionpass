#!/bin/bash -e
PRIVATEKEY=$(cat ../id_rsa)
go build -ldflags "-X 'github.com/fasmide/onionpass/ssh.RawPrivateKey=$PRIVATEKEY'"  ../

# move and reload systemd service
scp -P2222 onionpass.service root@onionpass.net:/etc/systemd/system
ssh -p2222 root@onionpass.net systemctl daemon-reload

# move old executable out of the way (dont fail)
ssh -p2222 root@onionpass.net 'mv /onionpass $(mktemp onionpass-update-XXXXX)' || true

# move the new executable
scp -P2222 onionpass root@onionpass.net:/onionpass

# allow binary to listen to privileged ports
ssh -p2222 root@onionpass.net setcap cap_net_bind_service=+ep /onionpass

# restart daemon
ssh -p2222 root@onionpass.net systemctl restart onionpass
sleep 0.5
ssh -p2222 root@onionpass.net systemctl status onionpass

rm onionpass