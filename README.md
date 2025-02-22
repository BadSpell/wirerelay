# WireRelay

WireRelay is a Go project that allows a WireGuard server located in a NAT environment to be accessed externally via a relay server with a public IP.  

```
+-----------------+    +----------------------+    +----------------------+    +---------------+
|    WG Server    | -> |  WireRelay (Client)  | -> |  WireRelay (Server)  | -> |   WG Client   |
|      (NAT)      |    |        (NAT)         |    |      (Public IP)     |    |     (NAT)     |
+-----------------+    +----------------------+    +----------------------+    +---------------+
```

## Build

```bash
git clone https://github.com/badspell/wirerelay
cd wirerelay
go build wirerelay.go
``` 
    

## Run
Use the `-server` or `-client` flag to select the mode.


### Server Mode
```
$ ./wirerelay -server 0.0.0.0:21804 -token Pa63OJB0uVMlDMjjajRLI75EwQYegPwYliIANVJTchU=
Registration token (base64): Pa63OJB0uVMlDMjjajRLI75EwQYegPwYliIANVJTchU=
Relay server listening on 0.0.0.0:21804
Registered relay client: 23.192.228.80:53311
```
The server displays a Base64 registration token, used for client mode.
If the `-token` option is omitted, a random token is generated.


### Client Mode
```
$ ./wirerelay -client public.server:21804 -target localhost:51820 -token Pa63OJB0uVMlDMjjajRLI75EwQYegPwYliIANVJTchU=
Registered success with relay server.
```
After connecting to the relay server, you'll see a success message and it will link to the target server (e.g., WireGuard).
