#!/bin/bash

cd /docker-project

docker compose down 
docker compose pull
docker compose up -d 