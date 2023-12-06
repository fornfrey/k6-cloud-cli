#!/bin/bash
set -eux
set -o pipefail

make build;

export K6_CLOUD_TOKEN=user-token;
export K6_CLOUD_HOST=http://localhost:9875;
export K6_CLOUD_API_HOST=http://api.dev.k6.io

./k6 cloud loadzone list

./k6 cloud organization list

./k6 cloud project list

./k6 cloud schedule delete $(./k6 cloud schedule list --org-id 3 | awk 'FNR == 3 {print $1}')
./k6 cloud schedule list --org-id 3
./k6 cloud schedule set 1 never
./k6 cloud schedule list --org-id 3

./k6 cloud test download 1
./k6 cloud test get 1
./k6 cloud test list

./k6 cloud testrun download 1
./k6 cloud testrun get 1
./k6 cloud testrun list 1
