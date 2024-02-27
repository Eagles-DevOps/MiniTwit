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
 
provider "digitalocean" {
  token = var.do_token
}

variable "docker_image_main" {
  description = "Docker image for the main application"
  type        = string
  default     = "/minitwit_main:latest"
}
 
resource "digitalocean_droplet" "main-app" {
  image  = "ubuntu-22-04-x64"
  name   = "main-app"
  region = "ams3"
  size   = "s-1vcpu-1gb"
  
  user_data = <<-EOF
    #!/bin/bash
    apt-get update
    apt-get install -y docker.io
    docker run -d -p 80:80 ${var.docker_image_main}
  EOF
}

output "droplet_ip_main_app" {
  value       = digitalocean_droplet.main-app.ipv4_address
  description = "The public IP address of the main-app droplet."
}

variable "docker_image_api" {
  description = "Docker image for the API"
  type        = string
  default     = "/minitwit_api:latest"
}

resource "digitalocean_droplet" "api" {
  image  = "ubuntu-22-04-x64"
  name   = "api"
  region = "ams3"
  size   = "s-1vcpu-1gb"
  
  user_data = <<-EOF
    #!/bin/bash
    apt-get update
    apt-get install -y docker.io
    docker run -d -p 80:80 ${var.docker_image_api}
  EOF
}

output "droplet_ip_api" {
  value       = digitalocean_droplet.api.ipv4_address 
  description = "The public IP address of the api droplet."
}
