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
  type        = string
}

variable "private_key" {}

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

  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = var.private_key
  }
}

resource "digitalocean_droplet" "api" {
  image  = "docker-20-04"
  name   = "api"
  region = "ams3"
  size   = "s-1vcpu-1gb"
  ssh_keys = [
    data.digitalocean_ssh_key.Viktoria_key.id
  ]

  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = var.private_key
  }
}

data "digitalocean_ssh_key" "Viktoria_key" {
  name = "Viktoria_key"
}