## Various instructions for general workflow

### Set up debugging in vs code:

To setup debugging create file .vscode/launch.json with the following content:

```
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch file",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceRoot}/go/minitwit.go"
        }
    ]
}
```

If vs-code is opened in Minitwit/ you should now be able to debug the program by pressing F5.
Make sure to run 'docker compose down' and exit any running 'go' commands beforehand to ensure the port is open.


## Testing API service
- Get the location of the service. Should be linked in main README.md file
- Paste the link into Postman (or other similar software)
- set Authentication to `Basic Auth` - username: `simulator` -pwd: `super_safe!`
- make a request and test


## Vagrant
- create API token in digital ocean.
- place it in the code
- generate SSH-KEY localy (guide in digital Ocean how to set this up)-- needs to be done only once
- reference the SSH-KEY name in the action
- run `vagrant up`
- Update the README.md if this is the production app with latest changes.

Be aware:
Vagrant may complain if go versions look like this 1.18.0 instead of 1.18.