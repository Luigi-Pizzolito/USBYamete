#!/bin/bash
cd "$(dirname "$0")"
kill `cat daemon.pid`
rm daemon.pid