# Epoxy

A very small proxy, it will try to do http requests or on getting an https or upgraded connection then TCP tunnel to the required destination.
The reason for this proxy is that given a header `X-Proxy-Hint` it can then route the traffic to the appropriate backconnect proxy for save costs and give us a higher rate of availability.

## How it Works
The proxy works by first inspecting the HTTP Method, if the Method is CONNECT then we know it will be either a WebSocket or HTTPS connection. 

### HTTP
If it is a basic HTTP connection then an HTTP transport is set up that will make an HTTP call to the remote and once a 200 is returned it will copy the body contents back onto the response for the original request. 

### HTTPS / WS
HTTPS and web sockets are a little bit trickier but imagine a simple TCP tunnel, and that is in essence what your client will connect over to the remote. The steps to create the tunnel are to receive the Client request, then create a new TCP connection to the remote and using two GO routines copy the data between the client and remote sources. 

## Running
To start the proxy locally you will need to build the proxy and then run

```
go build && ./epoxy start
```

Your proxy will then be running on `localhost:9001`

