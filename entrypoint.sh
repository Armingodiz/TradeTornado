#!/bin/bash

set -e

# Run migrations
./my-app migrate all

# Start the main application
./my-app "$@"
