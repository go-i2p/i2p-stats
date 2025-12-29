# i2p-stats

This tool gathers statistics from a running I2P router and uses them to generate various ways of organizing and viewing those statistics.
It is also a nasty, ugly hackjob I'm in the process of revising.

At this time, the primary way of viewing them is as a static web site, so that it can easily be viewed on an eepsite or shared via github pages.
It also generates a json file for every stat, which can be fetched by referencing the time and date it was gathered.
It is intended to be run about every ten minutes, as a cron job, and for the output to be shared on an eepsite.

## Cron Example:

```
*/10  * * * * $HOME/go/bin/i2p-stats -dir $HOME/go/src/github.com/eyedeekay/i2p-stats/weather
```

## Loop Example:

```
while true; do i2p-stats -dir=$(pwd)/weather && git add weather && git commit -am "checkin";  git push --all; sleep 10m; done
```