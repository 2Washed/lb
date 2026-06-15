* Load balancing is the process of distributing traffic on multiple servers to improve performance and reliability

## What are the types of lb?
- Static

```
Static lb algorithms distribute the workload on the servers without taking into consideration their current state
Pros: Quick to setup
Cons: Inefficient

Example Algorithms: Round robin, Weighted round robin, source-ip hash
```
- Dynamic
```
Dynamic lb algorithms take into consideration the current the server's current state (Health, Availability, Workload ...)

Example algorithms: least connection, weighted least connection
```

# Rate limiting

## What are the types or rate limiting ?
- User rate limiting: based on api key or ip address
- Geographic rate limiting: applied based on the geographical proximity
- Server rate limiting: rate limits depend on the server (size, network, bandwidth)

## rate limiting algorithms
- Token bucket
```
We have a bucket that fills out with tokens, each time we have a new request we consume a token, if there's no tokens left we drop the request
- Token bucket can be easily implemented with a counter.
- The token is initiated to zero.
- Each time a token is added, counter is incremented to 1.
- Each time a unit of data is sent, counter is decremented by 1.
- When the counter is zero, host cannot send data.
```
- Leaky bucket
```
The algorithm saves the incoming requests into a queue and discards all the requests after queue is full
1. Initialize a counter to n at the tick of the clock.
2. Repeat until n is smaller than the packet size of the packet at the head of the queue.

    Pop a packet out of the head of the queue, say P.
    Send the packet P, into the network
    Decrement the counter by the size of packet P. 

3. Reset the counter and go to step 1. 
```
- Fixed window
```
Tracks the number of requests in a fixed time window and resets the window when the time expires
Traditional implementations may allow bursts at the edge of the boundary
[ x  x   xx][xxx   x ] <- suppose we allow 4 / interval, this is technically allowed but it results in a burst of 5 reqs near the boundary
```
- Sliding / rolling window
```
Uses a continuously moving time frame to limit the number of requests, which helps distribute the requests evenly over time
```
- Sliding / rolling logs
```
Saves the request timestamps
For each new request:
- Remove timestamps older than the window.
- Count remaining timestamps.
- Allow if count < limit.
```

### Sources
- [Cloudflare's "What is load balancing?"](https://www.cloudflare.com/learning/performance/what-is-load-balancing/)
- [GeeksForGeeks's "Static load balancing algorithms"](https://www.geeksforgeeks.org/system-design/static-load-balancing-algorithms/)
- [Load balancing and rate limiting for dummies](https://medium.com/@lakshay.bhambri/load-balancing-and-rate-limiting-for-dummies-part-4-b1de9a6f0353)
- [GeeksForGeeks: rate limiting algorithms](https://www.geeksforgeeks.org/system-design/rate-limiting-algorithms-system-design/)