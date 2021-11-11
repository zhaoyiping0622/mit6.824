#!/bin/bash

if [ $# -lt 2 ]
then
  echo usage: $0 times tests
  exit 1
fi

packages=${*:2}
echo packages to test: $packages

if [ -e testout ] 
then 
  rm -r testout
fi

mkdir testout

for loop in `seq 1 $1`
do
  echo test round $loop
  out=testout/out$loop
  echo "SLOW=1 debug=raftapp,kvraft PORT=6063 go test -v -failfast $packages > $out"
  debug=raftapp,kvraft PORT=6063 go test -v -failfast $packages > $out
  if [ $? -ne 0 ] 
  then 
    echo test failed output in $out
  else
    echo test passed
  fi
done
