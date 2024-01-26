#! /bin/env bash

# Navigate to the parent directory of the current script.
#cd ..

# Execute the 'build' target from the Makefile.
# This will compile the 'scour' application.
make build

# Run the 'scour' application with the POST method.
# The request is made to "https://httpbin.org/post" with an "accept: application/json" header.
echo "Running POST"
./build/scour -X POST "https://httpbin.org/post" -H "accept: application/json" -v

# Run the 'scour' application with the PUT method.
# The request is made to "https://httpbin.org/put" with an "accept: application/json" header.
echo "Running PUT"
./build/scour -X PUT "https://httpbin.org/put" -H "accept: application/json" -v

# Run the 'scour' application with the GET method.
# The request is made to "https://httpbin.org/get" with an "accept: application/json" header.
echo "Running GET"
./build/scour -X GET "https://httpbin.org/get" -H "accept: application/json" -v

# Run the 'scour' application with the DELETE method.
# The request is made to "https://httpbin.org/delete" with an "accept: application/json" header.
echo "Running DELETE"
./build/scour -X DELETE "https://httpbin.org/delete" -H "accept: application/json" -v

# Run the 'scour' application with the PATCH method.
# The request is made to "https://httpbin.org/patch" with an "accept: application/json" header.
echo "Running PATCH"
./build/scour -X PATCH "https://httpbin.org/patch" -H "accept: application/json" -v
