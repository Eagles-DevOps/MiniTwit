# terraform for setting up k3s


Required secrets/tokens:

DO read/write token - for creating droplets
DO read/update token - for use with limited rights to read/update reserved ip in case of node going down.
Simply API key - for updating dns record


Self chosen:

k3s token - used between nodes for creating cluster
rancher pw - password for the admin user in web interface



When deploying it is assumed files called prod-secrets.env and test-secrets.env exists in config folder with key-value:
POSTGRES_PW=****


For convinience a scripts for deployment can be created be created at redeploy.sh
```
terraform plan -out main.tfplan -var='rancher-pw=****' -var='simply_api_key=****' \
    -var='k3s_token=****' -var='email=****' -var='do_read_token=****' \
    -var 'do_token=****' -var 'pvt_key=terranetes.pvt'
terraform apply "main.tfplan"
```