source ~/.bash_profile

cd /docker-project

docker compose down 
docker compose -f docker-compose.yml pull
docker compose -f docker-compose.yml up -d