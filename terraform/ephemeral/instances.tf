terraform {
  required_version = ">= 1.0.0"

  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

# variable "digital_ocean_token" {
#    description = "DigitalOcean API Token"
#    type = string
# }

variable "do_token" {
  description = "DigitalOcean API Token"
  type        = string
}

# variable "pvt_key_path" {}

variable "pvt_key" {}

provider "digitalocean" {
  token = var.do_token
  #   token = var.digital_ocean_token # change for the CI/CD pipeline
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
    # private_key = file(var.pvt_key_path)
    private_key = var.pvt_key
    timeout     = "2m"
  }

  # provisioner "remote-exec" {
  #   inline = [
  #     "export PATH=$PATH:/usr/bin",
  #     "sudo apt update",
  #     "sudo apt install -y curl",
  #     "curl -fsSL https://get.docker.com -o get-docker.sh",
  #     "sudo sh get-docker.sh",
  #     "sudo docker run hello-world"
  #   ]
  # }

  # provisioner "file" {
  #   source      = "provision.sh"
  #   destination = "/tmp/provision.sh"
  # }

  provisioner "file" {
    source      = "docker-compose.yml"
    destination = "/tmp/docker-compose.yml"
  }

  provisioner "remote-exec" {
    script = "provision.sh"
  }
}


# output "droplet_ip_main_app" {
#   value       = digitalocean_droplet.main-app.ipv4_address
#   description = "The public IP address of the main-app droplet."
# }

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
