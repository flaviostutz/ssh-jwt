#!/bin/bash

set -e
# set -x

echo "Starting SSH-JWT..."
ssh-jwt \
     --log-level=$LOG_LEVEL \
     --bind-host=$BIND_HOST \
     --port=$BIND_PORT \
     --enable-remote-forwarding=$ENABLE_REMOTE_FORWARDING \
     --enable-local-forwarding=$ENABLE_LOCAL_FORWARDING \
     --enable-pty=$ENABLE_PTY     

