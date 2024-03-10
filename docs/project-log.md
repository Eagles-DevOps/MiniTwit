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

## 29/02/2024 JAN
I'll try to do some `hello world` with terraform. Because I haven't worked with it before, I'll run *main.tf* on my digital ocean to see how it works. 

**how to install terraform (mac)**
https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli
~~~
brew install hashicorp/tap/terraform
~~~
**tutorials for terraform and digital ocean**
- 1h video: https://www.youtube.com/watch?v=dSJ6zenfRK8
	- I skipped parts from the second part of the video. But it's nice to know about variables, workspaces, parameters, creating multiple resources without count. maybe I'll get back to this when I have some time. 
- official terraform digital ocean docs https://www.digitalocean.com/community/tutorials/how-to-use-terraform-with-digitalocean

**my test token** (the space before the command removes it from bash history)
~~~
 export TF_VAR_do_token=dopXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
 export TF_VAR_pvt_key=$(cat ~/.ssh/terraform)
export TF_LOG=1 #don't use if not needed
~~~

**basic terraform commands**
~~~
alias tf=terraform # I use this alias because it's just quicker to write
tf init
tf plan
tf apply
tf destroy
~~~

Now I can finally, get to the reddit post I found earlier to solve my problem
https://www.reddit.com/r/Terraform/comments/ca5gvb/keep_floating_ip_after_destroy/
They propose this structure. One state in *persistent* folder and one in *ephemeral*
```
digitalocean/
	persistent/ 
		floating-ip.tf 
	ephemeral/ 
		instances.tf 
		data.tf
```

chatGPT explained what he meant in the comment about using state file from different workspace. 
```
# persistent workspace in digitalocean/persistent/ 
output "reserved_ip_address" { value = digitalocean_reserved_ip.example.ip_address }
(terraform init and terraform apply)

# ephemeral workspace in digitalocean/ephemeral/ 
data "terraform_remote_state" "other_workspace" {
  backend = "local"
  config = {
    path = "<path_to_other_workspace_state_file>"
  }
}

resource "digitalocean_reserved_ip_assignment" "example" {
  ip_address = data.terraform_remote_state.other_workspace.outputs.reserved_ip_address
  droplet_id = digitalocean_droplet.example.id
}
```


"By default, Terraform writes its state file to your local filesystem. This is okay for personal projects, but once you start working with a team, things get messy. In a team, you need to make sure everyone has an up to date version of the state file and ensure that two people aren’t making concurrent changes" 

We'll have two states. and we already have problem with just one. because it's not shared, let's fix that first. 

**Storing Terraform State in Digital Ocean Space**
this is the tutorial I read earlier: [dev.to-blog](https://dev.to/aleixmorgadas/storing-terraform-state-in-digital-ocean-space-3a97) looks good 

**or possible alternative, but much more complicated**
(might not be free anymore) [4 Reasons To Try HashiCorp’s (New) Free Terraform Remote State Storage](https://medium.com/runatlantis/4-reasons-to-try-hashicorps-new-free-terraform-remote-state-storage-b03f01bfd251) 

It's storage by the company behind terraform. In the blog they said it's free. don't know if that's still true. it's specifically for this purpose and it also has GUI.

update: I looked at it seems that it's great, but it will take some time to setup and also we'll have to create a new organization. I think it's overkill for what we need. 
It's not a priority so **I'll just make an issue for it.**

So I want back to the issue at hand and created the folders:
```
digitalocean/
	persistent/ 
		reserved-ip.tf 
	ephemeral/ 
		instances.tf 
```
the shared state from persistent workspace seems to be working without any problems.

**provision droplet with ssh**
good docs for digital ocean an terraform
- https://www.digitalocean.com/community/developer-center/how-to-run-terraform-on-digitalocean

we will need first and ssh key
https://docs.digitalocean.com/products/droplets/how-to/add-ssh-keys/create-with-openssh/
```
ssh-keygen -f ~/.ssh/terraform
```

**read public key to env**
~~~
export TF_VAR_pvt_key=$(cat ~/.ssh/terraform)
~~~

so it failed somehow it got stuck in a loop.
```
"digitalocean_reserved_ip_assignment.example" is waiting for "digitalocean_reserved_ip_assignment.example (expand)"


2024-02-29T17:21:28.208+0100 [TRACE] dag/walk: vertex "provider[\"registry.terraform.io/digitalocean/digitalocean\"] (close)" is waiting for "digitalocean_reserved_ip_assignment.example"

"digitalocean_reserved_ip_assignment.example (expand)" is waiting for "digitalocean_droplet.prod"
"digitalocean_reserved_ip_assignment.example (expand)" is waiting for "digitalocean_droplet.prod"

2024-02-29T17:21:33.210+0100 [TRACE] dag/walk: vertex "root" is waiting for "provider[\"registry.terraform.io/digitalocean/digitalocean\"] (close)"
2024-02-29T17:21:33.210+0100 [TRACE] dag/walk: vertex "digitalocean_reserved_ip_assignment.example" is waiting for "digitalocean_reserved_ip_assignment.example (expand)"
2024-02-29T17:21:33.210+0100 [TRACE] dag/walk: vertex "provider[\"registry.terraform.io/digitalocean/digitalocean\"] (close)" is waiting for "digitalocean_reserved_ip_assignment.example"
```

After refactoring this doesn't seem to be an issue anymore. Hopefully it won't show up again later. 

**change visibility for the packages**
I had to change visibility for the packages, because they were not public. 
- to do that, I had to change policy in Github organization to allow public packages.
- there is an option to change the visibility at the very bottom of the package page.

~~~
# testing if I can pull the packages now that they are public
docker pull ghcr.io/eagles-devops/minitwit:build-to-dockerhub
docker pull ghcr.io/eagles-devops/minitwit-web-app:251d1993f2c781321a20abf14c6513db391d624f
~~~

**how to ssh with key**
ssh -i ~/.ssh/terraform -o StrictHostKeyChecking=no root@188.166.201.66

don't forget to change to your digital ocean key
**local development with terraform
~~~
 export TF_VAR_do_token=dop_v1_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
 export TF_VAR_pvt_key=$(cat ~/.ssh/terraform)
alias tf=terraform
alias qq="terraform destroy -auto-approve ; terraform apply -auto-approve"
alias ss="ssh -i ~/.ssh/terraform -o StrictHostKeyChecking=no root@188.166.201.66"
~~~

**I've deleted the old packages and changed to a shorter naming convention** (group/package)
- docker pull ghcr.io/eagles-devops/app:latest
- docker pull ghcr.io/eagles-devops/api:latest

## 07/03/2024 JAN
**our logs are not persistent I'll look at some options to fix that**
https://volkovlabs.io/blog/nginx-loki-grafana-20230129/

https://ruanbekker.medium.com/logging-with-docker-promtail-and-grafana-loki-d920fd790ca8

https://abhiraj2001.medium.com/monitoring-docker-containers-with-grafana-loki-and-promtail-4302a9417c0d

Off topic: I found this command that will automatically create new branch if they don't exist. 
```bash
git config --global --add --bool push.autoSetupRemote true
```

## 08/03/2024 JAN
The db file "minitwit.db" used to be in the root folder "minitwit-api". This was all good until we required to have it persistent even if the container gets deleted. 

In docker there are 3 ways to make something persistent so that it survives the recreation of a container. In docker compose all the options are under "volumes:" which might be confusing, because all of them work a bit differently.

**Passing a Host File to a Docker Container**:
~~~
version: '3.8'
services:
  file-reader:
    image: alpine
    volumes:
      - ./hostfile.txt:/containerfile.txt

~~~
In this example, `hostfile.txt` from the host machine is mounted into the container as `containerfile.txt`. Any changes made to `containerfile.txt` will be reflected in `hostfile.txt` on the host machine. This would be our first option, but the problem was that the file needs to exist before the container is started. In our case the go application creates the db file, so was can't use this option.

**Passing a Host Directory to a Docker Container**
~~~
version: '3.8'
services:
  directory-reader:
    image: alpine
    volumes:
      - ./hostdirectory:/container/dir1/dir2

~~~

Here, the `hostdirectory` from the host machine is mounted into the container as `containerdirectory`. Any files within `hostdirectory` on the host machine will be accessible inside the container at `containerdirectory`. The difference is that if the directory doesn't exist on the host it will be created automatically. That applies also to the container and the whole path is created. If the mounted path was /data/dir1/dir2 all 3 folders would just appear in the container as defined by the path. Also if ./hostdirectory doesn't exist it will be created. 

It would be better to call those first 2 options mounts or binds. The next option is the actual docker volumes, which we don't use. 

**Using Docker Volume to Persist a Directory Inside a Container**:
~~~
version: '3.8'
services:
  data-processor:
    image: myappimage
    volumes:
      - mydata:/app/data

volumes:
  mydata:
~~~

This docker compose file creates a volume called mydata. You can think of volume as a virtual disk that you connect to a docker container. The volume has a mount point in this case /app/data that you can access the files from inside the container. 
you can move, delete, export docker volumes with docker volume xxx commands. Volumes seem great why I personally avoid them? they are difficult to access from the host. Only the container has access to the data inside the volume. 

**in our app**
~~~
volumes:
	- ./sqlitedb-app/:/usr/src/app/sqlitedb/
~~~
this will create an empty folder (if doesn't exists) called sqlitedb-app in the directory on the host where the docker-compose file is located. It will also create empty folder (if doesn't exists) in /usr/src/app/sqlitedb/ which is the home directory of the app. 

when the container is started it will look for the folder and create the db file inside /usr/src/app/sqlitedb/ which from the go perspective is just ./sqlitedb (because when the image was created we used the WORKDIR which is basically like cd. 
~~~
FROM golang:1.22

WORKDIR /usr/src/app      <------ [this line]


# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change

COPY *.go go.mod go.sum ./

RUN go mod download && go mod verify
... 
~~~

when the app creates the the db file it will also show up on the host. but for the host it will be in the folder ./sqlitedb-api/


now if you run it locally, docker doesn't create sqlite directory. (those lines didn't work until today)
```
if err == nil {
		fmt.Println("directory of the database exists")
	} else if os.IsNotExist(err) {
		fmt.Println("directory of the database does not exist, will create new one")
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("Fatal Error: creating directory for db: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Println("db directory created")
		}
	}>)
```

now there is one more feature we added that is this line.

```
dbPath := os.Getenv("SQLITEPATH")
if len(dbPath) == 0 {
	dbPath = "./sqlite/minitwit.db"
}
```

you can override the dbPath if you want to.
for example, if you had a test database with some other data you could just use the command. 
export SQLITEPATH=./mytestdb.db
a if you run the go api it will use this is path instead.

in docker compose you can also set env variables. 
~~~
environment:
- SQLITEPATH=/usr/src/app/sqlitedb/minitwit.db
~~~

but this is redundant (I probably I should've just deleted that) if I don't specify the SQLITEPATH it will be empty. And if it's empty it will be set to ./minitwit.db and since the working dir of the go api is /usr/src/app/sqlitedb/ is the same as ./minitwit.db

**how does ./sqlitedb-api get created inside the VM?**

if you specify a path for a mount in docker-compose that doesn't exist. docker compose will automatically create it. 

**why do we use different paths locally versus in docker?**
"./sqlite/minitwit.db" Or dbPath := os.Getenv("SQLITEPATH") (the latter returns empty string when run locally)

The relative paths are actually the same it should be just "./sqlite/minitwit.db" for both local development and running in the container. We just need to be careful to always start the app from the folder where the minitwit-api.go file is located. 

I'll just remove this from our docker-compose 
```
environment:
	- SQLITEPATH=/usr/src/app/sqlitedb/minitwit.db

