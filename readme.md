# Epoxy

A very small proxy, it will try to do http requests or on getting an https or upgraded connection then TCP tunnel to the required destination.
The reason for this proxy is that given a header `X-Proxy-Hint` it can then route the traffic to the appropriate backconnect proxy for save costs and give us a higher rate of availability.

## How it Works
The proxy works by first inspecting the HTTP Method, if the Method is CONNECT then we know it will be either a WebSocket or HTTPS connection. 

### HTTP
If it is a basic HTTP connection then an HTTP transport is set up that will make an HTTP call to the remote and once a 200 is returned it will copy the body contents back onto the response for the original request. 

### HTTPS / WS
HTTPS and web sockets are a little bit trickier but imagine a simple TCP tunnel, and that is in essence what your client will connect over to the remote. The steps to create the tunnel are to receive the Client request, then create a new TCP connection to the remote and using two GO routines copy the data between the client and remote sources. 

## Directing the Proxy Selection 
You are able to provide hints to the proxy to enable a better selection of proxies for your needs. To do this you can provide a header `X-Proxy-Hint` which will guide the selection, this header is just a small json payload

```
ie.
{ 
  "use" : "192.168.1.1:80"
}
```

| name     | type     | description                                                                                              |
| -------- | -------- | -------------------------------------------------------------------------------------------------------- |
| use      | string   | provide the address of the proxy to use                                                                  |
| failed   | []string | an array of strings for proxies that you dont want to use                                                |
| priority | int      | select a group of proxies to use, 1 being the highest ranked group and any number lower is well... lower |

The idea behind the priority is that Proxies that are more expensive will get a higher priority and so you will have to specify that you want to use these. They should be reserved to instances where the data is more critical and we know the cheaper proxies will fail.

By default the proxy priority is set to 2, so you should get the middle tier proxies by default.

## Running
### Configuration
You will need to set up the configuration file to get the proxy to run
The file is located at `config/[env.yml]` however you can always override the `config/default.yml` too

| name      | type           | description                                                    |
| --------- | -------------- | -------------------------------------------------------------- |
| port      | int            | port the proxy will run on                                     |
| proxies   | map[string]int | the list of backconnect proxy uri's followed by their priority |
| blacklist | []string       | regex list of uri endpoints to block traffic to                |

### Environment variables
To use environment variables they must be prefixed with `EPOXY_`, this will mark it as an Epoxy variable.

Any configuration variable can be overriden by an environment variable, you will need to replace the `-` in the environment variable with `_`. 

ie 
```
epoxy start --port 80 
```
will overide the environment variable

```
export EPOXY_PORT=9010

```
will override the configuration file
```
port: 9001
```

### Setting Proxies and Blacklist
Since both of these are yaml structures any attempt to override these in the commndline will also need to be in the yaml format. You can use the Short Notation to achieve this

```
export EPOXY_PROXIES={\"1.1.1.1:80\":2, \"2.2.2.2:80\":2}
export EPOXY_BLACKLIST="[\"test.com\", \"foo.com\"]"
```


### Starting
To start the proxy locally you will need to build the proxy and then run

```
go build && ./epoxy start
```

Your proxy will then be running on `localhost:9001`


## Thanks
This project is built using the following libraries:

[Cobra](https://github.com/spf13/cobra)

[Viper](https://github.com/spf13/viper)

[logrus](github.com/sirupsen/logrus)

