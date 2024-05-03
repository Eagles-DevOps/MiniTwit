#!/bin/bash

echo $SHELL

      sleep 10
      export PATH=$PATH:/usr/bin

      fallocate -l 4G /swapfile
      chmod 600 /swapfile
      mkswap /swapfile
      swapon /swapfile
      echo "/swapfile swap swap defaults 0 0" >> /etc/fstab


        sleep 60
      echo "Install K3s"
      curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="v1.28.8+k3s1" sh -s - server --server https://${NODE1_IP}:6443 --token=${K3S_TOKEN}
      sudo ufw allow 6443


       echo "{\"DIGITALOCEAN_TOKEN\":\"${DIGITALOCEAN_TOKEN}\", \"RESERVED_IP\":\"${RESERVED_IP}\", \"DROPLET_ID\":\"${DROPLET_ID}\" , \"PRIMARY_NODE\":\"node1\"}" >> /tmp/env
chmod +x /tmp/update_ip.sh
systemctl enable --now ip.timer
      echo DONE