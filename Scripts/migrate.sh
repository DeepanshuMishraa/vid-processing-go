#!/bin/bash

# 1. Load DATABASE_URL from .env file
# This command ignores comments and empty lines, then exports the variables
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
else
    echo ".env file not found"
    exit 1
fi

# 2. Verify DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL is not set in .env"
    exit 1
fi

# 3. Handle Migration Commands
COMMAND=$1
SHIFT_COUNT=$2

case $COMMAND in
    up)
        migrate -path ./migrations -database "$DATABASE_URL" up $SHIFT_COUNT
        ;;
    down)
        # Default to 1 to prevent accidental full database wipe
        COUNT=${SHIFT_COUNT:-1}
        migrate -path ./migrations -database "$DATABASE_URL" down $COUNT
        ;;
    force)
        migrate -path ./migrations -database "$DATABASE_URL" force $SHIFT_COUNT
        ;;
    *)
        echo "Usage: ./migrate.sh {up|down|force} [version/count]"
        exit 1
        ;;
esac
