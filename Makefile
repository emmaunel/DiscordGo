SERVER=d2Server
LIN_SERVER=lind2Server
DIRECTORY=bin
MAC=macos-agent
LINUX=linux-agent
WIN=windows-agent.exe
RASP=rasp
BSD=bsd-agent
FLAGS=-ldflags "-s -w"
WIN-FLAGS=-ldflags -H=windowsgui


all: clean create-directory agent-mac agent-linux agent-windows agent-rasp

create-directory:
	mkdir ${DIRECTORY}

server-mac:
	echo "Compiling mac server"
	env GOOS=darwin GOARCH=amd64 go build -o ${DIRECTORY}/${SERVER} cmd/server/main.go

server-linux:
	echo "Compiling linux server"
	env GOOS=linux GOARCH=amd64 go build -o ${DIRECTORY}/${LIN_SERVER} cmd/server/main.go

agent-mac:
	echo "Compiling macos binary"
	env GOOS=darwin GOARCH=amd64 go build ${FLAGS} -o ${DIRECTORY}/${MAC} cmd/agent/main.go

agent-linux:
	echo "Compiling Linux binary"
	env GOOS=linux GOARCH=amd64 go build ${FLAGS} -o ${DIRECTORY}/${LINUX} cmd/agent/main.go

agent-windows:
	echo "Compiling Windows binary"
	env GOOS=windows GOARCH=amd64 go build ${WIN-FLAGS} -o ${DIRECTORY}/${WIN} cmd/agent/main.go

agent-rasp:
	echo "Compiling RASPI binary"
	env GOOS=linux GOARCH=arm GOARM=7 go build ${FLAGS} -o ${DIRECTORY}/${RASP} cmd/agent/main.go

agent-fuckbsd:
	echo "Compiling FUCKBSD binary"
	env GOOS=freebsd GOARCH=amd64 go build ${FLAGS} -o ${DIRECTORY}/${BSD} cmd/agent/main.go

clean:
	rm -rf ${DIRECTORY}
