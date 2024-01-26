#! /bin/env bash

cd ..
make build

echo "Running POST"
./build/scour -X POST "https://httpbin.org/post" -H "accept: application/json" -v

echo "Running PUT"
./build/scour -X PUT "https://httpbin.org/put" -H "accept: application/json"

echo "Running GET"
./build/scour -X GET "https://httpbin.org/get" -H "accept: application/json"

echo "Running DELETE"
./build/scour -X DELETE "https://httpbin.org/delete" -H "accept: application/json"

echo "Running PATCH"
./build/scour -X PATCH "https://httpbin.org/delete" -H "accept: application/json"

