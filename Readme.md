# TCP Tic Tac Toe

A peer to peer TCP tic tac toe game in plain Golang.

## How to play

make sure you have go installed on your machine.

1. run the server `go run main.go`

2. once the server is running, open two terminal windows and connect to the server using `telnet localhost 8080`

3. start playing!

## playing over the local network

1. run `ifconfig` or `ip addr` on Linux/MacOS or `ipconfig` on Windows.

2. find `en0` and find `inet 192.168.X.X`, than copy ip address and paste it to 196 line.

3. run the server `go run main.go`

4. once the server is running, open two terminal and connect to the server using `telnet {your ip from 196 line} 8080`

## playing over the internet

if you want to play with a friend over the internet, you can use a service like [ngrok](https://ngrok.com/)
to expose your local server to the internet.

1. run the server `go run server.go`

2. run ngrok `ngrok tcp 8080`

3. from another machines, connect to the ngrok address using `telnet <ngrok_address> <ngrok_port>`

4. start playing!