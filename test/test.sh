#!/bin/bash
assert() {
  expected="$1"
  input="$2"

  ../gocc "$input" > tmp.s
  gcc -o tmp tmp.s
  ./tmp
  actual="$?"

  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo "$input => $expected expected, but got $actual"
    exit 1
  fi
}

assert 3 "3"
assert 4 "2+2"
assert 5 "2+4-1"

echo OK
