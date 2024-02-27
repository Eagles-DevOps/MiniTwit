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
 
resource "digitalocean_droplet" "main-app" {
  image  = "ubuntu-22-04-x64"
  name   = "main-app"
  region = "ams3"
  size   = "s-1vcpu-1gb"

  connection {
    type        = "ssh"
    user        = "root"
    private_key = file("~/.ssh/id_rsa")  # Path to your SSH private key
  }

  provisioner "remote-exec" {
    inline = [
      "sudo apt update",
      "sudo apt install -y golang",
      "mkdir -p /usr/src/app",
      "cd /usr/src/app",
      "git clone https://github.com/Eagles-DevOps/MiniTwit.git",
      "cd MiniTwit/minitwit-web-app",
      "go run minitwit.go",  # Assuming your Go files are named main.go
    ]
  }
}


output "droplet_ip_main_app" {
  value       = digitalocean_droplet.main-app.ipv4_address
  description = "The public IP address of the droplet."
}

resource "digitalocean_droplet" "api" {
  image  = "ubuntu-22-04-x64"
  name   = "main-app"
  region = "ams3"
  size   = "s-1vcpu-1gb"

  connection {
    type        = "ssh"
    user        = "root"
    private_key = file("~/.ssh/id_rsa")  # Path to your SSH private key
  }

  provisioner "remote-exec" {
    inline = [
      "sudo apt update",
      "sudo apt install -y golang",
      "mkdir -p /usr/src/app",
      "cd /usr/src/app",
      "git clone https://github.com/Eagles-DevOps/MiniTwit.git",
      "cd MiniTwit/minitwit-api",
      "go run minitwit-api.go",  # Assuming your Go files are named main.go
    ]
  }
}

output "droplet_ip_api" {
  value       = digitalocean_droplet.api.ipv4_address 
  description = "The public IP address of the droplet."
}