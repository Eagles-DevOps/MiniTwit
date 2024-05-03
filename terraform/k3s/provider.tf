terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~>3.0"
    }
  }
}

variable "do_token" {}
variable "do_read_token" {}
variable "pvt_key" {}
variable "rancher-pw" {}
variable "k3s_token" {}
variable "email" {}
variable "simply_api_key" {}



provider "digitalocean" {
  token = var.do_token
}

data "digitalocean_ssh_key" "terranetes" {
  name = "terranetes"
}