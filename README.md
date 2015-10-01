#Slurp Server
This is a simple file server implementation, I got the specifications for creating this program from this [distributed system paper](http://www.cs.columbia.edu/~roxana/teaching/DistributedSystemsF12/labs/), they specified c++ as the implementation language but the point of this exercise was to get practiced at my golang skills. The other half of this is [slurp client](github.com:/ebuckley/slurp_client) it speaks the required protocol to send files over the wire.

The exercise specifications are [here](http://www.cs.columbia.edu/~roxana/teaching/DistributedSystemsF12/labs/lab0.html)

#Usage

```
$ ./slurp_server /data 0.0.0.0:1337
```
The preceding example shows us how to invoke the built slurp_server to serve the `/data` folder where it expects the slurp client to know the servers host port.

#Features
- LRU cache implementation
- safe closedown
- good logging (subjectivly)

#Tests
I tested the LRU implementation, but nothing else is automatically tested (I'm a terrible terrible person I know)

```
#from the project directory
$ cd LRU
$ go test
PASS
ok  	github.com/ebuckley/slurp_server/LRU	0.004s

```
