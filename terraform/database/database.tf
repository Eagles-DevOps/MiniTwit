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

resource "digitalocean_database_cluster" "postgres-cluster" {
  name       = "minitwit-db-cluster"
  engine     = "pg"
  version    = "16"
  size       = "db-s-1vcpu-1gb"
  region     = "ams3"
  node_count = 1
}

resource "digitalocean_database_db" "database-example" {
  cluster_id = digitalocean_database_cluster.postgres-cluster.id
  name       = "minitwit-db"
}

