go get github.com/nlopes/slack
go get github.com/lib/pq
go build -o bin/application application.go
cp ./appSettings.json bin
