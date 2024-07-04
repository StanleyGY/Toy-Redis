
# Motivation

Inspired by Redis's fast performance and determined to learn more about the Redis internal, I decided to embark on this personal project to understand what design makes Redis to stand out among its peer databases. This Toy Redis aims to recreate some of the basic query functionalities using similar data structures and architectural designs adopted by the actual Redis server. Although the original server is written in C, this project is implemented in Golang to leverage its simplicity and ease of use the language offers.

Redis's single threaded design is known to improve its impressive performance, by avoiding thread context switching and lock contention. Toy Redis recreates the single-threaded design, using `epoll` to accept and listen for the incoming client connections.

Redis server utilizes intricately designed data structures to speed up various query performance. Toy Redis mimics to implement complex data structures to improve server performance in various places:
- Radix tree is optimal for string lookup with shared prefix
    + Fast lookup of stream entries where id is represented in unix timestamp
    + Fast lookup of client timeouts
- Skip list is optimal for sorted sets
    + Faster query performance on sorted set

Redis makes use of an event-driven architecture where each request, after its execution, generates an event to be processed in the event queue. In Toy Redis, the semi-event-driven design decouples the server module from `cmdexec`, making it easier to support blocked APIs that might generate delayed responses (e.g. `XREAD`).

Toy Redis is an ongoing personal effort to recreate the very basics of Redis server. To focus on the core elements of Redis, many assumptions have been made to ease the development, so there are lots of unhandled corner cases. For example:
- What if client makes another call while it is supposed to block wait for its prev command to finish?
- How to handle connection draining when server is killed?

Also, there are many basic features that are lacking (but on TODO list), for example:
- Redis replication
- Redis sentinel
- Redis persistence

# Supported Commands

#### Server Commands

| Command | Purpose | Note |
|---|---|---|
| PING | Server replies "pong" |
| ECHO message | Server replies with user-supplied message |

#### Simple Set Commands

| Command | Purpose | Note |
|---|---|---|
| SET key value [NX] [EX secs \| PX millisecs] | Set a value in simple dict |
| GET key | Get a value in simple dict |

#### Sorted Set Commands

| Command | Purpose | Note |
|---|---|---|
| ZADD key [NX] score member [score member ...] | Add a member in sorted set |
| ZREM key member [member ...] | Remove a member in sorted set |
| ZSCORE key member | Return the score of a member |
| ZCOUNT key min max | Count number of members with scores between min and max |
| ZRANGEBYSCORE key min max [WITHSCORES]| Return all members with scores between min and max |
| ZRANK key member [WITHSCORE] | Return the rank of a member |
| ZRANGE key start stop [WITHSCORES] | Return all members with ranks between start and stop |

#### Stream Commands

| Command | Purpose | Note |
|---|---|---|
|XADD key <*> field value [field value ...]| Add entries to a stream at key, and return the stream ID
|XRANGE key start end [COUNT count]| Query entries with stream IDs between start and end |
|XREAD [COUNT count] [BLOCK milliseconds] STREAMS key [key...] id [id...]| Read entries since provided stream IDs at multiple keys. Blocking wait if entries do not exist until timeout occurs.

#### Geo-Spatial Commands

| Command | Purpose | Note |
|---|---|---|
|GEOADD key longitude latitude member [longitude latitude member ...]| Add members with lats and lons
|GEODIST key member1 member2 [M \| KM]| Calculate haversine distance between member1 and member2
|GEOHASH key member [member ...]| Compute geo-hash of a member
|GEORADIUS key longitude latitude radius| Query members that are within radius of the provided coordinate | Unoptimized. Redis uses skiplist and geo bounding box for optimized search query

# Run Server

To start a server, simply execute the bash script `./spawn_redis_server.sh`.

To interact with the server, use `redis-cli`.