#!/bin/sh

# Set the default values for the configuration options
SERVICE_TYPE="yaml"
SERVICES_FILE="services.yaml"
REDIS_NAMESPACE="frontman"
REDIS_URI="redis://localhost:6379"
API_ADDR="0.0.0.0:8080"
GATEWAY_ADDR="0.0.0.0:8000"
LOG_LEVEL="info"

# Check if environment variables are set and override the defaults
if [ -n "$FRONTMAN_SERVICE_TYPE" ]; then
    SERVICE_TYPE="$FRONTMAN_SERVICE_TYPE"
fi

if [ -n "$FRONTMAN_SERVICES_FILE" ]; then
    SERVICES_FILE="$FRONTMAN_SERVICES_FILE"
fi

if [ -n "$FRONTMAN_REDIS_NAMESPACE" ]; then
    REDIS_NAMESPACE="$FRONTMAN_REDIS_NAMESPACE"
fi

if [ -n "$FRONTMAN_REDIS_URI" ]; then
    REDIS_URI="$FRONTMAN_REDIS_URI"
fi

if [ -n "$FRONTMAN_API_ADDR" ]; then
    API_ADDR="$FRONTMAN_API_ADDR"
fi

if [ -n "$FRONTMAN_GATEWAY_ADDR" ]; then
    GATEWAY_ADDR="$FRONTMAN_GATEWAY_ADDR"
fi

if [ -n "$FRONTMAN_LOG_LEVEL" ]; then
    LOG_LEVEL="$FRONTMAN_LOG_LEVEL"
fi

# Get the full path of the frontman.yaml file
CONFIG_PATH="/app/frontman.yaml"

# Replace the configuration options in the YAML file
sed -i "s|SERVICE_TYPE|$SERVICE_TYPE|g" "$CONFIG_PATH"
sed -i "s|SERVICES_FILE|$SERVICES_FILE|g" "$CONFIG_PATH"
sed -i "s|REDIS_NAMESPACE|$REDIS_NAMESPACE|g" "$CONFIG_PATH"
sed -i "s|REDIS_URI|$REDIS_URI|g" "$CONFIG_PATH"
sed -i "s|API_ADDR|$API_ADDR|g" "$CONFIG_PATH"
sed -i "s|GATEWAY_ADDR|$GATEWAY_ADDR|g" "$CONFIG_PATH"
sed -i "s|LOG_LEVEL|$LOG_LEVEL|g" "$CONFIG_PATH"


./frontman