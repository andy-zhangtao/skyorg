# skyorg
Connect public network and private network


## Design

```go
|-----------------|                               |-----------------|
|   skyServer     | <-----Tcp Socket A----------->|   skyClient     |
|-----------------|                               |-----------------|
                                                           |
                                                           | Tcp Socket B
	                                                      \|/
	                                              |-----------------|
	                                              |  local server   |
	                                              |-----------------|
```

SkyServer will manager a tcp socket pool. When SkyClient connect SkyServer, they will create N(10 as default) tcp connect.

If there is a user connect SkyServer, SkyServer will choose a idle connect for waitting request. If user send request, SkyServer will receives this request, and forward to SkyClient. Then SkyClient will create a real tcp connect between SkyClient with Local Server. 

Tcp Socket A will forward Tcp Socket B to user. 