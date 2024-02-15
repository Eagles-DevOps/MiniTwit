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