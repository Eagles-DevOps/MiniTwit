#!/bin/bash

export PATH=$PATH:/usr/bin
cd /tmp/
docker compose down
docker compose pull
docker compose up -d