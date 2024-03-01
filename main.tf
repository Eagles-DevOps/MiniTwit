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

resource "digitalocean_droplet" "api" {
  image  = "docker-20-04"
  name   = "api"
  region = "ams3"
  size   = "s-1vcpu-1gb"
  ssh_keys = [
    data.digitalocean_ssh_key.Viktoria_key.id
  ]
}

data "digitalocean_ssh_key" "Viktoria_key" {
  name = "Viktoria_key"
}

provisioner "file" {
  source = "deploy.sh"
  destination = "/docker-project/deploy.sh"
}

provisioner "file" {
  source = "docker_compose.yml"
  destination = "/docker-project/docker_compose.yml"
}

provisioner "remote-exec" {
  inline = [
    "chmod 777 /tmp/deploy.sh"
    "/docker-project/deploy.sh"
  ]
}