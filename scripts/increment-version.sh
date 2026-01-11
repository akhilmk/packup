#!/bin/bash

VERSION_FILE="VERSION"

if [ ! -f "$VERSION_FILE" ]; then
    echo "0.0.1" > "$VERSION_FILE"
    echo "0.0.1"
    exit 0
fi

VERSION=$(cat "$VERSION_FILE")

# Split version into major.minor.patch
IFS='.' read -r -a parts <<< "$VERSION"
MAJOR="${parts[0]}"
MINOR="${parts[1]}"
PATCH="${parts[2]}"

# Increment patch version
PATCH=$((PATCH + 1))

NEW_VERSION="$MAJOR.$MINOR.$PATCH"
echo "$NEW_VERSION" > "$VERSION_FILE"
echo "$NEW_VERSION"
