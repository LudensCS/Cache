#!A:/Git/bin/bash
trap 'rm server.exe;kill 0' EXIT

go build -o server.exe main.go

./server -load true

./server -port 8001 &
./server -port 8002 &
./server -port 8003 -api true &

sleep 2

echo ">>> start test"

curl "http://localhost:9999/api?key=Jack" &
curl "http://localhost:9999/api?key=Lucy" &
curl "http://localhost:9999/api?key=David" &
curl "http://localhost:9999/api?key=Jack" &

wait
