#!/bin/bash
kill -9 `ps -ef|grep server_main|grep -v grep|awk '{print $2}'`
