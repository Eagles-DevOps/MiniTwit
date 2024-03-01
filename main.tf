terraform {
  required_version = ">= 1.0.0"
 
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

variable "do_token" {
   description = "DigitalOcean API Token"
   type = string
}
 
provider "digitalocean" {
  token = var.do_token
}
 
resource "digitalocean_droplet" "app" {
  image  = "docker-20-04"
  name   = "app"
  region = "ams3"
  size   = "s-1vcpu-1gb"
  ssh_keys = [
    data.digitalocean_ssh_key.Viktoria_key.id
  ]
}

output "droplet_ip_main_app" {
  value       = digitalocean_droplet.main-app.ipv4_address
  description = "The public IP address of the main-app droplet."
}

resource "digitalocean_droplet" "api" {
  image  = "docker-20-04"
  name   = "api"
  region = "ams3"
  size   = "s-1vcpu-1gb"
  ssh_keys = [
    data.digitalocean_ssh_key.Viktoria_key.id
  ]
}

output "droplet_ip_api" {
  value       = digitalocean_droplet.api.ipv4_address 
  description = "The public IP address of the api droplet."
}

data "digitalocean_ssh_key" "Viktoria_key" {
  name = "Viktoria_key"
}

provisioner "file" {
  source = "deploy.sh"
  destination = "/tmp/deploy.sh/"
}

provisioner "file" {
  source = "docker_compose.yml"
  destination = "/tmp/docker_compose.yml/"
}