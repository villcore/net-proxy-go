base=`dirname $0`
echo `$base/stop-server.sh`
echo "$base/server_main"
echo `nohup $base/server_main > server.out 2>&1 &`
