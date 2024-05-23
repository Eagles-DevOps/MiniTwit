## CI-CD


### PR tests
A complete description of stages and tools included in the CI/CD chains, including deployment and release of your systems.

When code changes related to an issue are ready to be merged in to main, a pull request is opened, and unit- and integration tests are then run as a GitHub Action. These - along with a review, and passing the quality gate of the static analysis tool, are a prerequisit for the pull request to be completed.


### Deployment
When the pull request is completed, the changed are merged to main triggering the ci-cd workflow. 
- Build docker image
- Push docker image to registry
- Deploy to K3S with Helm



### Other workflow

We have a few other workflows as part of our setup.

As required a workflow generating a PDF top the report directory has been created.

A manually triggered workflow for running linters exists in linters.yaml.

Lastly a workflow for automating a weekly release is running at the end of each week.
At this point it would make more sense to consider each automated deploy a release instead, but the automated release can be seen as a weekly snapshot.
