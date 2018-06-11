# About

This container runs a script that invokes puic-poll in an infinite loop such that puic-poll
makes 4 runs with 256 requests each waiting for 1s to 2s after a request has completed. 

# Build

You'll need go at least version 1.9.1 and you need to have the `mami-project/plus-quic-go` (which is a fork)
und the import path `lucas-clemente/quic-go`. If you have the dependencies set up correctly invoke `./build.sh`
which compiles the `puic-poll.go` and creates the binary under `files/puic-poll` and then continues with
building the docker image. 



# Test locally

Run for example:

```
sudo docker run -it puic-poll eth0 "https://172.17.0.1:6121/data/256;https://172.17.0.1:6121/data/128MiB;https://172.17.0.1:6121/16MiB"
```

The first argument is the interface to listen on and the second argument is a list of URLs to poll. 
