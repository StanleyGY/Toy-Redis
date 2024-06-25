#!/bin/bash
run() {
    echo $@ | redis-cli
}

echo "Test cmd ZADD: adding members"
run ZADD myset 1 A 2 B 3 C 4 D

echo "Test cmd ZADD: adding duplicates"
run ZADD myset 1 A 2 B

echo "Test cmd ZADD: updating score"
run ZADD myset 10 A 20 B 30 C

echo "Test cmd ZADD: updating score with NX"
run ZADD myset NX 40 D

echo "Test cmd ZCOUNT"
run ZCOUNT myset 20 30

echo "Test cmd ZSCORE"
run ZSCORE myset A
run ZSCORE myset C

echo "Test cmd ZRANGEBYSCORE"
run ZRANGEBYSCORE myset 10 40 WITHSCORES
run ZRANGEBYSCORE myset 3 4

echo "Test cmd ZRANK"
run ZRANK myset A
run ZRANK myset D

echo "Test cmd ZRANGE"
run ZRANGE myset 0 4

echo "Test cmd ZREM: removing from non-existing key"
run ZREM fake A

echo "Test cmd ZREM: removing members"
run ZREM myset A B E
