AGENT=d2
SERVER=d2Server
DIRECTORY=bin


all: clean create-directory server agent-mac agent-linux agent-windows

create-directory:
	mkdir ${DIRECTORY}

server:
	echo "Compiling server"
	go build -o ${DIRECTORY}/${SERVER} cmd/server/main.go

agent-mac:
	echo "Compiling macos binary"
	env GOOS=darwin GOARCH=amd64 go build -o ${DIRECTORY}/macos-agent cmd/agent/main.go

agent-linux:
	echo "Compiling Linux binary"
	env GOOS=linux GOARCH=amd64 go build -o ${DIRECTORY}/linux-agent cmd/agent/main.go

agent-windows:
	echo "Compiling Windows binary"
	env GOOS=windows GOARCH=amd64 go build -o ${DIRECTORY}/windows-agent cmd/agent/main.go

clean:
	rm -rf ${DIRECTORY}
