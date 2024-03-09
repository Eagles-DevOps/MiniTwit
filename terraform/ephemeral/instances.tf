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

variable "pvt_key" {}

provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_droplet" "prod" {
  image  = "docker-20-04"
  name   = "prod"
  region = "ams3"
  size   = "s-1vcpu-1gb"
  ssh_keys = [
    data.digitalocean_ssh_key.terraform.id
  ]

  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = var.pvt_key
    timeout     = "2m"
  }

  provisioner "file" {
    source      = "./config"
    destination = "/tmp/config"
  }

  provisioner "file" {
    source      = "~/terraform.env"
    destination = "/root/terraform.env"
  }

  provisioner "file" {
    source      = "docker-compose.yml"
    destination = "/tmp/docker-compose.yml"
  }

  provisioner "file" {
    source      = "provision.sh"
    destination = "/tmp/provision.sh"
  }

  provisioner "remote-exec" {
  inline = [
    "chmod +x /tmp/provision.sh",  # Ensure the script is executable
    "/tmp/provision.sh"            # Run the script
  ]
  }
}

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

data "digitalocean_ssh_key" "terraform" {
  name = "terraform"
}
