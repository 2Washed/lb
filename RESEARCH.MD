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
### Sources
- [Cloudflare's "What is load balancing?"](https://www.cloudflare.com/learning/performance/what-is-load-balancing/)
- [GeeksForGeeks's "Static load balancing algorithms"](https://www.geeksforgeeks.org/system-design/static-load-balancing-algorithms/)