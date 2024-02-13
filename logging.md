# Reflections



## Session02

Steps taken: 

Docker running: 

cd go/
docker build -t <name/name> .
docker run <name/name>


##
2024/02/13
dangr:
Added docker compose as well as converting refactored python tests to unit tests and containerizing them.

Can now be run with:
docker compose up -d --build
<<wait a few seconds>>
docker logs minitwit-tests-1

tests are failing as of now (to be expected)

