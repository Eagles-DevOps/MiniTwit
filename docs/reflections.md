# Reflections
MSc students reflections (to check in the future):
https://github.com/search?q=repo%3Aitu-devops%2Flecture_notes%20MSc%20students&type=code

### Programming language (Session 2)
choice: **GO**
considered: Java, C#

**GO**
- Lower memory consumption compared to ruby, Java etc.
- Scales well, which is important for us considering the userbase will increase over time.
- Easy to learn.
- Compiles to a single binary which makes it easy to deploy (compared to).
- Comes with an opinionated formatter (gofmt) ensuring code style consistency. This reduces discussions about styling (which does not bring value to the product).
- It is a language with high developer interest meaning that it will hopefully be ease future recruitment when the application sky-rockets in popularity.

### Software Artifacts
choice: **docker**
considered: VMs, linux Packages, LXC, go packages
- lower overhead compared to VMs
- supported on most linux distributions regardless of package managers
- containers isolate the environment from the host system
- support for using different language compared to language specific artifacts
- support micro-services in our case allow us to run api and app with docker compose
- community support

### CI/CD Pipelines Tool
choice: **GitHub Actions**
considered: Jenkins, GitLab CI/CD, Bamboo

**GitHub Actions**
- already integrated into code repository of our choice (Github)
- minimal setup, compared to tools such as Jenkins
- runs on cloud without need of provisioning
- modern and easy to use UI

### Artifact registry
choice: **DockerHub**
considered: GitHub Packages
- largest docker registry

### Infrastructure automation platforms
choice: **terraform**
previously used: vagrant

**Terraform**
- there is currently larger community behind terraform than vagrant
- less unexpected behavior we experienced compared to using vagrant