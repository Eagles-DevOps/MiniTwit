## Logging


## How do you monitor your systems and what precicely do you monitor?

For monitoring we use Prometheus. We do so by incrementing gauges or vectors whenever an event has succesfully occured.
In the system we monitor a multitude of things, for business data we log:
    - We monitor the amount of users getting created.
    - The amount of new followers on the platform
    - Amount of new messages posted
    - The total amount of reads and writes made to the database between releases.
Besides incrementing counters we also monitor back-end data:
    - The amount of failed database read-writes
    - Whether there is a connection to the database
    - Succesful HTTP requests

Monitoring these gives us an insight to the extend of traffic passing through our API.

## What do you log in your systems and how do you aggregate logs?

??? 







### Vizualisation