terraform {
  required_version = ">= 1.0.0"
 
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

# variable "digital_ocean_token" {
#    description = "DigitalOcean API Token"
#    type = string
# }
 
variable "do_token" {
    description = "DigitalOcean API Token"
    type = string
}
 
provider "digitalocean" {
    token = var.do_token
#   token = var.digital_ocean_token # change for the CI/CD pipeline
}
 
resource "digitalocean_droplet" "prod" {
  image  = "ubuntu-22-04-x64"
  name   = "prod"
  region = "ams3"
  size   = "s-1vcpu-1gb"
}

# output "droplet_ip_main_app" {
#   value       = digitalocean_droplet.main-app.ipv4_address
#   description = "The public IP address of the main-app droplet."
# }

### Add the static IP

data "terraform_remote_state" "other_workspace" {
  backend = "local"

  config = {
    path = "../persistent/terraform.tfstate"
  }
}

resource "digitalocean_reserved_ip_assignment" "example" {
  ip_address = data.terraform_remote_state.other_workspace.outputs.reserved_ip_address
  droplet_id = digitalocean_droplet.prod.id
}