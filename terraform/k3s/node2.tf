

resource "digitalocean_droplet" "node2" {
  image = "ubuntu-20-04-x64"
  name = "node2"
  region = "fra1"
  size = "s-2vcpu-2gb"
  ssh_keys = [
    data.digitalocean_ssh_key.terranetes.id
  ]
  provisioner "file" {
  source = "./node2etcd.sh"
  destination = "./node2.sh"
  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = file(var.pvt_key)
    timeout = "2m"
  }
  }

  provisioner "file" {
  source = "./ip.service"
  destination = "/etc/systemd/system/ip.service"
  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = file(var.pvt_key)
    timeout = "2m"
  }
  }

  provisioner "file" {
  source = "./ip.timer"
  destination = "/etc/systemd/system/ip.timer"
  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = file(var.pvt_key)
    timeout = "2m"
  }
  }

  provisioner "file" {
  source = "./update_ip.sh"
  destination = "/tmp/update_ip.sh"
  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = file(var.pvt_key)
    timeout = "2m"
  }
  }
  
  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = file(var.pvt_key)
    timeout = "2m"
  }
}
resource "null_resource" "remote_exec2" {
  depends_on = [digitalocean_droplet.node2]

  # Connection block to use the Floating IP for SSH connection
  provisioner "remote-exec" {
    connection {
      type = "ssh"
      user = "root"
      private_key = file(var.pvt_key)
      host        = digitalocean_droplet.node2.ipv4_address
    }

    inline = [
    "chmod +x node2.sh",
    "DIGITALOCEAN_TOKEN=${var.do_read_token} RESERVED_IP=${digitalocean_floating_ip.public_ip.ip_address} DROPLET_ID=${digitalocean_droplet.node2.id} RANCHER_PW=${var.rancher-pw} NODE1_IP=${digitalocean_droplet.node1.ipv4_address_private} NODE2_IP=${digitalocean_droplet.node2.ipv4_address_private} K3S_TOKEN=${var.k3s_token} ./node2.sh",
    "rm node2.sh"
  ]
  }
}