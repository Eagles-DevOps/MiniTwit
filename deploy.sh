#!/bin/bash

export PATH=$PATH:/usr/bin
cd /docker-project

docker compose down 
docker compose up -d 
docker compose pull