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


When skyclient connect skyserver success, there will be three connections. 

+ Message Conn

    The real data message connection. All request will transfer in this connection.
    
+ Control Conn

    A connecion wil transfer control message. In this connection, skyServer will ask skyClient creates specify amount Message Connctions.  
    
+ Proxy Conn

    The connection between SkyClient and Local Server. When SkyClient receives data from Message Connection, it will transfer data to local server within this connection.
    
    
Three connections as shown in the following figure:

```go
                            
                     |-------------------------|
---Request --------->| skyServer(0.0.0.0:33333)|------------
                     |-------------------------|           |Control Connection (0.0.0.0:33335)
	                            |                          |
	                            |Message Connection        |
	                            |                          |
                     |-------------------------|           |
	                 | skyClient(0.0.0.0:33334)| <---------|
	                 |-------------------------|
	                 	        |
	                 	        |Proxy Connection
	                 |-------------------------|
	                 |      Local Server       |
	                 |-------------------------|
```        

So skyServer will host three ports:

+ request port 
    
    Default 33333. Receive user request. 
    
+ data message port

    Default 33334. Transfer request data between skyServer with skyClient
    
+ control port

    Default 33335. Transfer control data        