#!/bin/bash

echo $SHELL

sleep 10
export PATH=$PATH:/usr/bin

###############
# Create swap #
###############

fallocate -l 4G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo "/swapfile swap swap defaults 0 0" >>/etc/fstab


###############
# Install K3S #
###############

echo "Install K3s"
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="v1.28.8+k3s1" sh -s - server --token=${K3S_TOKEN} --cluster-init --tls-san=${NODE1_IP} --tls-san=${RESERVED_IP} --tls-san=eagles.danielgron.dk

sudo ufw allow 6443
mkdir /kube
sleep 5
kubectl config view --raw >>/kube/config
export KUBECONFIG=/kube/config
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
helm repo add rancher-latest https://releases.rancher.com/server-charts/latest

sudo kubectl create namespace cattle-system
sudo kubectl create namespace cert-manager

sudo kubectl --kubeconfig="/kube/config" apply -f https://github.com/jetstack/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# Maximum number of retries
max_retries=10

# Count of current retry
retry_count=0

# Function to execute the command and check the output
function check_command {
    output=$(sudo kubectl --kubeconfig="/kube/config" get pods --namespace cert-manager) # Execute the command and capture the output
    count=$(echo "$output" | grep -o '1/1' | wc -l) 

    if [ $count -ge 3 ]; then
        echo "Success: Found 'running' three times or more."
        echo $output
        return 0 # Success
    else
        echo "Retry #$retry_count: 'running' found $count times."
        echo $output
        return 1 # Fail
    fi
}

# Main retry loop
while [ $retry_count -lt $max_retries ]; do
    if check_command; then
        sleep 10
        break # Exit loop if successful
    fi

    ((retry_count++))
    sleep 20
done

if [ $retry_count -eq $max_retries ]; then
    echo "Failed: Maximum retries reached without success."
fi


###################
# Install Rancher #
###################

helm install rancher rancher-latest/rancher --kubeconfig="/kube/config" --namespace cattle-system --set hostname=rancher.eagles.danielgron.dk --set bootstrapPassword=${RANCHER_PW} --set replicas=1 --set ingress.tls.source=letsEncrypt --set letsEncrypt.email=${EMAIL} --set letsEncrypt.environment=production

echo "{\"DIGITALOCEAN_TOKEN\":\"${DIGITALOCEAN_TOKEN}\", \"RESERVED_IP\":\"${RESERVED_IP}\", \"DROPLET_ID\":\"${DROPLET_ID}\" , \"PRIMARY_NODE\":\"node1\"}" >>/tmp/env
chmod +x /tmp/update_ip.sh
systemctl enable --now ip.timer

echo DONE setting up k3s

######################
# Create DNS records #
######################

SIMPLY="https://api.simply.com/2/my/products/danielgron.dk/dns/records"
MINITWIT_API_RECORD="{\"type\": \"A\",\"name\": \"minitwit-api.danielgron.dk\", \"data\": \"${RESERVED_IP}\", \"priority\": 10, \"ttl\": 600}"
MINITWIT_APP_RECORD="{\"type\": \"A\",\"name\": \"minitwit-app.danielgron.dk\", \"data\": \"${RESERVED_IP}\", \"priority\": 10, \"ttl\": 600}"
MINITWIT_TEST_API_RECORD="{\"type\": \"A\",\"name\": \"minitwit-api-test.danielgron.dk\", \"data\": \"${RESERVED_IP}\", \"priority\": 10, \"ttl\": 600}"
MINITWIT_TEST_APP_RECORD="{\"type\": \"A\",\"name\": \"minitwit-app-test.danielgron.dk\", \"data\": \"${RESERVED_IP}\", \"priority\": 10, \"ttl\": 600}"
RANCHER_RECORD="{\"type\": \"A\",\"name\": \"rancher.eagles.danielgron.dk\", \"data\": \"${RESERVED_IP}\", \"priority\": 10, \"ttl\": 600}"

curl --user UE355473:$SIMPLY_API_KEY -d "$MINITWIT_API_RECORD" -H "Content-Type: application/json" -X POST $SIMPLY
curl --user UE355473:$SIMPLY_API_KEY -d "$MINITWIT_APP_RECORD" -H "Content-Type: application/json" -X POST $SIMPLY
curl --user UE355473:$SIMPLY_API_KEY -d "$MINITWIT_TEST_API_RECORD" -H "Content-Type: application/json" -X POST $SIMPLY
curl --user UE355473:$SIMPLY_API_KEY -d "$MINITWIT_TEST_APP_RECORD" -H "Content-Type: application/json" -X POST $SIMPLY
curl --user UE355473:$SIMPLY_API_KEY -d "$RANCHER_RECORD" -H "Content-Type: application/json" -X POST $SIMPLY

function check_ip {
    output=$(ping -c 1 rancher.eagles.danielgron.dk)         # Execute the command and capture the output
    count=$(echo "$output" | grep -o ${RESERVED_IP} | wc -l) 

    if [ $count -ge 1 ]; then
        echo "Success: Ip has been updated for dns"
        echo $output
        return 0 # Success
    else
        echo "Retry #$retry_count: Ip not found"
        echo $output
        return 1 # Fail
    fi
}

# Wait for DNS to propagate - can take a while for digital ocean
while [ $retry_count -lt 100 ]; do
    if check_ip; then
        break # Exit loop if successful
    fi

    ((retry_count++)) # Increment retry count
    sleep 120
done

#############################
# Create initial deployment #
#############################

PROD_NS=minitwit-prod
TEST_NS=minitwit-test

kubectl apply -f /yaml/clusterrole.yaml
kubectl create serviceaccount github-actions
kubectl create clusterrolebinding continous-deployment --clusterrole=continuous-deployment --serviceaccount=default:github-actions

sudo kubectl create namespace $PROD_NS
kubectl apply -n $PROD_NS -f <(envsubst </yaml/letsencrypt.yaml)

kubectl apply -f /yaml/ingress.yaml -n $PROD_NS
kubectl apply -f /yaml/middleware.yaml -n $PROD_NS

kubectl create configmap api --from-env-file=/config/prod-config.env -n $PROD_NS
kubectl create secret generic api --from-env-file=/config/prod-secrets.env -n $PROD_NS
helm upgrade --install minitwit-api /Charts/minitwit-api -f /Charts/minitwit-api/values.yaml --set image.tag=latest -n $PROD_NS
rm /config/prod-secrets.env

sleep 60

sudo kubectl create namespace $TEST_NS
kubectl apply -f /yaml/test.postgres.yaml -n $TEST_NS
kubectl create configmap api --from-env-file=/config/test-config.env -n $TEST_NS
kubectl create secret generic api --from-env-file=/config/test-secrets.env -n $TEST_NS
rm /config/test-secrets.env

helm upgrade --install minitwit-api /Charts/minitwit-api -f /Charts/minitwit-api/values.yaml --set image.tag=latest -n $TEST_NS
