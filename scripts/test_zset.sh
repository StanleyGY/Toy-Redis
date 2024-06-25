#!/bin/bash
run() {
    echo $@ | redis-cli
}

echo "Test adding members"
run ZADD myset 1 A 2 B 3 C 4 D

echo "Test adding duplicates"
run ZADD myset 1 A 2 B

echo "Test updating score"
run ZADD myset 10 A 20 B 30 C

echo "Test updating score with NX"
run ZADD myset NX 40 D

echo "Test cmd ZCOUNT"
run ZCOUNT myset 20 30

echo "Test cmd ZRANGEBYSCORE"
run ZRANGEBYSCORE myset 10 40 WITHSCORES
run ZRANGEBYSCORE myset 3 4

echo "Testing removing from non-existing key"
run ZREM fake A

echo "Test removing members"
run ZREM myset A B E

echo "Test getting score"
run ZSCORE myset A
run ZSCORE myset C
