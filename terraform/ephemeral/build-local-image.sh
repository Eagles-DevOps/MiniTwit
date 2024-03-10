#!/bin/bash
echo "shadowing ghcr.io/eagles-devops/api:latest with local image.."

# get the directory where the script is located
script_dir=$(dirname "$0")
cd "$script_dir"/../../minitwit-api/
echo "changed directory to: $(pwd)"
ls
docker build -t ghcr.io/eagles-devops/api:latest .
