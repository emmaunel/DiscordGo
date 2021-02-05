AGENT=d2
SERVER=d2Server
DIRECTORY=bin
MAC=macos-agent
LINUX=linux-agent
WIN=windows-agent
FLAGS=-ldflags "-s -w"


all: clean create-directory server agent-mac agent-linux agent-windows

create-directory:
	mkdir ${DIRECTORY}

server:
	echo "Compiling server"
	go build -o ${DIRECTORY}/${SERVER} cmd/server/main.go

agent-mac:
	echo "Compiling macos binary"
	env GOOS=darwin GOARCH=amd64 go build ${FLAGS} -o ${DIRECTORY}/${MAC} cmd/agent/main.go

agent-linux:
	echo "Compiling Linux binary"
	env GOOS=linux GOARCH=amd64 go build ${FLAGS} -o ${DIRECTORY}/${LINUX} cmd/agent/main.go

agent-windows:
	echo "Compiling Windows binary"
	env GOOS=windows GOARCH=amd64 go build ${FLAGS} -o ${DIRECTORY}/${WIN} cmd/agent/main.go

clean:
	rm -rf ${DIRECTORY}
