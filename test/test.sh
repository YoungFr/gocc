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

assert 0 "0"
assert 1 "0 + 1"
assert 2 "2 - 0"
assert 4 "1 + 1 + 2"
assert 5 "(3+3) - 1"
assert 6 "2 * (30 - 28) + (2 / 1)"
assert 7 "((1+2) * 3) - (3 - (1/1))"
assert 8 "(2+1) * (3+1) / (14 / 7) + (4-2)"

echo OK
