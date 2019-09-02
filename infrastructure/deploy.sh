#!/bin/bash -e

go build ../

# move old executable out of the way
ssh -p2222 root@116.203.213.190 'mv /root/onionpass $(mktemp onionpass-update-XXXXX)'

# Move the new executable
scp -P2222 onionpass root@116.203.213.190:onionpass

ssh -p2222 root@116.203.213.190 systemctl restart onionpass
sleep 0.5
ssh -p2222 root@116.203.213.190 systemctl status onionpass