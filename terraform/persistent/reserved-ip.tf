terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

variable "do_token" {}

provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_reserved_ip" "prod-ip" {
  region = "ams3"
}

output "reserved_ip_address" {
  value = digitalocean_reserved_ip.prod-ip.ip_address
}

# data "digitalocean_account" "my_account_info" {
# }

# output "droplet_limit"{
#   value = data.digitalocean_account.my_account_info.droplet_limit
# }

# resource "digitalocean_droplet" "web" {
#   image  = "ubuntu-22-04-x64"
#   name   = "web-1"
#   region = "ams3"
#   size   = "s-1vcpu-1gb"
# }

# output "server_ip" {
#   value = digitalocean_droplet.web.ipv4_address
# }