#!/bin/bash

set -eu 
set -o pipefail

configure_rsyslog
configure_db "${DB}"

go run github.com/onsi/ginkgo/v2/ginkgo -keep-going -trace -r -fail-on-pending -randomize-all -p -race -timeout 20m "$@"

