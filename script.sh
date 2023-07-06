
#!/bin/bash

echo "------------$(date +%F' '%T)------------"
# 开发重启脚本

file=map-websocket

getPid() {
  docmd=$(ps aux | grep ${file} | grep ${file} | grep -v 'grep' | grep -v '\.sh' | awk '{print $2}')
  echo $docmd
}

start() {
  pidstr=$(getPid)
  if [ -n "$pidstr" ]; then
    echo "running with pids $pidstr"
  else
    rm -rf $file
    echo "正在编译中..."
    go build -o $file
    sleep 0.5
    printf "\n"
    printf "正在执行启动...稍候"
    printf "\n"
    nohup ./$file >../logs/map-server.txt 2>&1 &
    pidstr=$(getPid)
    echo "start with pids $pidstr Successful"
  fi
}

stop() {

  pidstr=$(getPid)
  if [ ! -n "$pidstr" ]; then
    echo "Not Executed!"
    return
  fi

  echo "kill $pidstr done"
  kill $pidstr

}

restart() {
  stop
  start
}

case "$1" in
start)
  start
  ;;
stop)
  stop
  ;;
restart)
  restart
  ;;
getpid)
  getPid
  ;;
esac