#!/bin/bash

export PATH=$PATH:/usr/bin
cd /tmp/

docker compose down 
docker compose up -d 
docker compose pull