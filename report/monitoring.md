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
For ease of access to the monitored data and for visualization, the group uses Grafanas dashboards, see ![Grafana Business data monitoring](/images/BusinessData.png)

## What do you log in your systems and how do you aggregate logs?


We log **Insert what we log** using *****
The aggregation of logs are done using *Loki* 





