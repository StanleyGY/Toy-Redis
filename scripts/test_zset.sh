#!/bin/bash
run() {
    echo $@ | redis-cli
}

echo "Test adding members"
run ZADD myset 1 A 2 B 3 C 4 D

echo "Test no duplicates"
run ZADD myset 1 A 2 B

echo "Test updating score"
run ZADD myset 10 A 20 B 30 C

echo "Test updating score with NX"
run ZADD myset NX 40 D

echo "Test removing members"
run ZREM myset A