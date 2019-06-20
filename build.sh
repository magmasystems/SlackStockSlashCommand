go get github.com/nlopes/slack
go get github.com/lib/pq
go get github.com/magmasystems/SlackStockSlashCommand
go build -o bin/application application.go
cp ./appSettings.json bin
