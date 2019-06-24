go get -u github.com/nlopes/slack
go get -u github.com/lib/pq
go get -u github.com/magmasystems/SlackStockSlashCommand
go build -o bin/application application.go
cp ./appSettings.json bin
