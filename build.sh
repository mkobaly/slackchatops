#build chatops for linux and windows
env GOOS=linux GOARCH=amd64 go build -o ./bin/chatops ./cmd/chatops
go build -o ./bin/chatops.exe ./cmd/chatops
