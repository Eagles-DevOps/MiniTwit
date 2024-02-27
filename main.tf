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
 
resource "digitalocean_droplet" "main-app-terraform " {
  image  = "ubuntu-22-04-x64"
  name   = "main-app"
  region = "ams3"
  size   = "s-1vcpu-1gb"
}

resource "digitalocean_droplet" "api" {
  image  = "ubuntu-22-04-x64"
  name   = "API-service"
  region = "ams3"
  size   = "s-1vcpu-1gb"
}

output "droplet_ip_main_app" {
  value       = digitalocean_droplet.main-app.ipv4_address
  description = "The public IP address of the droplet."
}

output "droplet_ip_api" {
  value       = digitalocean_droplet.api.ipv4_address 
  description = "The public IP address of the droplet."
}