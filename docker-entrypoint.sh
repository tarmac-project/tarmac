#!/bin/bash
## Entrypoint script for tarmac. This script is to show how to write 
## an entrypoint script that actually passes down signals from Docker.

## Load our DB Password into a runtime only Environment Variable
if [ -f /run/secrets/password ]
then
  echo "Loading DB password from secrets file"
  APP_DB_PASSWORD=$(cat /run/secrets/password)
  export APP_DB_PASSWORD
fi

## Run the Application
exec /app/tarmac/tarmac
