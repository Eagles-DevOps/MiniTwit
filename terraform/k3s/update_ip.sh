#!bin/bash

DROPLET_ID=$(cat /tmp/env | jq -r .DROPLET_ID)
RESERVED_IP=$(cat /tmp/env | jq -r .RESERVED_IP)
DIGITALOCEAN_TOKEN=$(cat /tmp/env | jq -r .DIGITALOCEAN_TOKEN)
PRIMARY_NODE=$(cat /tmp/env | jq -r .PRIMARY_NODE)


# Call DO API and get the value of the droplets id from the json comming back
CURRENT_ID=$(curl -X GET -s https://api.digitalocean.com/v2/reserved_ips/$RESERVED_IP     -H "Content-Type: application/json"     -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" | jq .reserved_ip.droplet.id)

#If the current droplet registered is this one we will not check further
if $CURRENT_ID == $DROPLET_ID; then
    exit 0;
fi

# Check if the currrent primary master is up and running
if kubectl get nodes | grep $PRIMARY_NODE | grep 'Ready'; then

exit 0;

# If node is not ready change reserved IP to this instance    
else
echo "Changing IP as ${PRIMARY_NODE} is not ready"

    curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
  -d '{"type":"assign","droplet_id":'"${DROPLET_ID}"'}' \
  "https://api.digitalocean.com/v2/reserved_ips/$RESERVED_IP/actions"

  #Important: requires node name to be the same for droplet and k3s

  PRIMARY_NODE==$(curl -X GET -s https://api.digitalocean.com/v2/reserved_ips/$RESERVED_IP     -H "Content-Type: application/json"     -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" | jq .reserved_ip.droplet.name)
  jq '.PRIMARY_NODE = '${PRIMARY_NODE}' /tmp/env > temp.json && mv temp.json /tmp/env
  echo "Primary node is now ${PRIMARY_NODE}"
fi


