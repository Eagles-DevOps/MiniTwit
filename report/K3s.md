## K3S
The API is now running on a leightweight Kubernetes cluster - k3s. This cluster spans two Server nodes. 
The cluster is spun up from scratch using terraform, and the infrastructure takes about an hour to spin up, as it needs to wait for dns propagation to be able to confirm domain ownership for the SSL certificate.
Configuration and secrets are deployed in the cluster as part of the setup process.

![secret](images/secret.png)


Rancher is running on top to provide a nice UI for management.

Letsencrypt is used for SSL certificates and is automatically created/renewed for deployments.


![ssl](images/ssl.png)


