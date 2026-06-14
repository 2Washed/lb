# Work Plan
## Version 0
- Simple backend proxy
```
- Listen to a port (8080)
- Forward every request to a single backend server
- Return the response to the client
```

## Version 1
- Simple load balancer
```
- Add support for multiple servers
- Use round robin algorithm
```

## Version 1.5

- Concurrency safe round robin
```
2 Solutions:
- mutex, Pros: Simple, Good for complex logic. Cons: Blocks other requests, more overhead
- atomic int, Pros: Lock-free (nonblocking), Very fast. Cons: Only works for single variables, harder to extends to complex logic
```

## Version 2
- Load balancer that only routes to healthy servers
```
- use goroutines to periodically check if servers are healthy
- only forward requests to healthy servers
```

## Version 3 
- A load balancer with failover
```
- Detect backend failures during request forwarding
- Retry another healthy backend
- Configure maximum retry attempts
```

## Version 4
- A load balancer that supports configuration files
```
- Use json or yaml to define load balancer configuration (opted to use json since it's already in the standard lib)
```

## Version 5
- A load balancer that supports weighted round robin
```
- Add a weight field and take it into consideration when picking a server from the list
- If no weight is specified in the configuration file, default to 1 (normal round robin)
```

## Version 6
- A load balancer that keeps track of server metrics
```
Metrics to keep track of:
- errors
- requests served
- active connections

Load balancer metrics:
- uptime
- total requests
```