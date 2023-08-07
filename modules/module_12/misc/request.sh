#!/bin/sh

# Function to generate a random number between 0 and 40
generate_random_number() {
  od -An -N2 -i /dev/urandom | awk -v max=41 '{print $1 % max}'
}


# Number of requests to send
total_requests=100

# URL to request
url="https://cloudnative.io/fibo"

# Loop to send requests
i=1
while [ "$i" -le "$total_requests" ]; do
  random_number=$(generate_random_number)
  request_url="${url}?n=${random_number}"
  echo "Sending request ${i}: ${request_url}"
  # Use curl to make the HTTP request
  curl -k "${request_url}"
  i=$((i + 1))
done
