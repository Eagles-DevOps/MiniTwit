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