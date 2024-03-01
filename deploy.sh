#!/bin/bash

cd /docker-project

docker compose down 
docker compose up -d 
docker compose pull