#!/bin/bash

# This script is run by an included GitHub Workflow, which expects it to exist at this location.

docker compose -f 'docker/compose.yml' build
