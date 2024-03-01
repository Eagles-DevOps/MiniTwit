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

variable "private_key" {
   description = "Private Key"
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

  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = var.private_key
  }

  provisioner "file" {
    source = "deploy.sh"
    destination = "/tmp/deploy.sh"
  }

  provisioner "file" {
    source = "docker_compose.yml"
    destination = "/tmp/docker_compose.yml"
  }

  provisioner "remote-exec" {
  inline = [
    "chmod +x /tmp/deploy.sh",
    "/tmp/deploy.sh"
  ]
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

  provisioner "file" {
    source = "deploy.sh"
    destination = "/tmp/deploy.sh"
  }

  provisioner "file" {
    source = "docker_compose.yml"
    destination = "/tmp/docker_compose.yml"
  }

  provisioner "remote-exec" {
  inline = [
    "chmod +x /tmp/deploy.sh",
    "/tmp/deploy.sh"
  ]
  }
}

data "digitalocean_ssh_key" "Viktoria_key" {
  name = "Viktoria_key"
}