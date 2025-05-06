# KMem Gateway Service

The Gateway Service is the central entry point for the KMem system, handling request routing, load balancing, and caching for all client interactions with the backend services.

# flow

http request -> caching -> routing -> protocol change(http -> grpc) -> to services
