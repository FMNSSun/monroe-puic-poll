# monroe-puic-poll

Docker container for Monroe experiments.

Configuration options (these need to be entered as key value pairs
in the monroe interface when scheduling):

 * URLs string: list of urls delimited by semi-colon.
 * WaitFrom int: minimum delay to wait after a request was completed to make the next request.
 * WaitTo int: maximum delay to wait after a request was completed to make the next request.
 * Collect int: how many requests to make per run and group together in a single log file.
 * IFaceName string: Name of the interface. This will bind the local address to the IP assigned to the interface on startup. Use `*` to specificy listening on all.
 * Runs int: how many runs to do. One run equals one log file.

The pull URL is `docker.io/munt/monroe-puic-poll`. Example config is:

```
"Collect":1024,"Runs":4,"URLs":"https://12.131.112.18:1010/data/4KiB","WaitFrom":10,"WaitTo":20,"IFaceName":"*"
```

## Caveats

Results are periodically synced from `/monroe/results` within the docker container to the monroe servers. To avoid "over-syncing"
it's recommended in the manual to write log files to a temporary directory and then move them to `/monroe/results`. This however has
the caveat that if the container is killed the log files obviously can't have been moved to `/monroe/results`. `monroe-puic-pull`
moves each run's log file after the run is completed to the `/monroe/results` directory (and then repeat with the next run). This means
that if the container is killed in the middle of the run, the current run's log file is lost (but log files from past runs
will be synced). This means that you need to select an appropriate amount of time and an appropriate data quota such that
your runs will complete with these settings (or that at least a few runs can complete before the container is killed).

## Local tests

To test the created docker container locally, you need to create a directory (e.g. `temp`) which contains a `results` directory and the
expected parameters in a JSON file called `config`. For example:
```
{
    "Collect":1024,
    "Runs":4,
    "URLs":"https://12.131.112.18:1010/data/4KiB",
    "WaitFrom":10,
    "WaitTo":20,
    "IFaceName":"*"
}
```

Now you can run docker using this command: `docker run -v <full path to temp directory>:/monroe monroe-puic-poll`.