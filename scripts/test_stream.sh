#!/bin/bash
run() {
    echo "$@" | redis-cli
}

run XADD mystream "*" A 1 B 1 C 1
run XADD mystream "*" A 2 B 2
run XADD mystream "*" A 3 B 3
run XADD mystream "*" A 4 B 4
run XADD mystream "*" A 5 B 5
run XADD mystream "*" A 6 B 6
echo "XRANGE mystream - +"
run XRANGE mystream - +
echo "XRANGE mystream - + 3"
run XRANGE mystream - + 3

echo "Test XREAD"
run XADD s1 "*" A1 1 A2 1
run XADD s1 "*" A3 1 A4 1
run XADD s2 "*" B1 2 B2 2
run XADD s3 "*" C1 3 C2 3
run XREAD COUNT 1 BLOCK 1000 STREAMS s1 s2 s3 0 0-0 1-0
run XREAD COUNT 1 BLOCK 1000 STREAMS s1 9999999999999
run XREAD COUNT 2 BLOCK 1000 STREAMS s1 s2 s3 0 0-0 9999999999999
