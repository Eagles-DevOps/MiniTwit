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


## 22/02/2024
DigitalOcean was picked due to its free trial we have had access to. As well DO provides an easy learning curve and easy set up. UI is simple and clean and no specialised knowledge is needed to spin up droplets.

## 27/02/2024

I moved the terraform files from root level into its own folder *iac* (Infrastructre as Code). The .gitignore was added to remove following from the git history:

* **tfstate file** - It is generally best practise not to store this in version control.

* **.terraform/** - This is dynamically created when using terraform locally. It can be specific to each developer environment.

* **other files** recommended by Github's gitignore

Propably we need to eventually find some kind of storage provideder (maybe [Space Object Storage](https://cloud.digitalocean.com/spaces)?). where we will maintain this file(s) and pull/push it from the GH action. However since our infrastructure (server) is propably pretty static right now, maybe we should not spend time looking at this right now.