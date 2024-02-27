terraform {
  required_version = ">= 1.0.0"
 
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

variable "digital_ocean_token" {
   description = "DigitalOcean API Token"
   type = string
}
 
provider "digitalocean" {
  token = var.digital_ocean_token
}
 
resource "digitalocean_droplet" "main-app-terraform" {
  image  = "ubuntu-22-04-x64"
  name   = "main-app"
  region = "ams3"
  size   = "s-1vcpu-1gb"
}

output "droplet_ip_main_app" {
  value       = digitalocean_droplet.main-app-terraform.ipv4_address
  description = "The public IP address of the droplet."
}