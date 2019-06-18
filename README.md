# Slack Stock Slash Command

Since leaving the corporate workforce and starting CTO as a Service, I have been slowly learning some things that have been on my TODO list for a while. Not having full-time management duties frees up your time, and every day, I find that there is so much more to learn. So, as I wind my way down the TODO list, I figure that I would start documenting some of my learnings so that it might be of use to others.

Even though I have been a Chief Architect and CTO for the last 15 years, I have still kept myself very technical, and I still code for pleasure, and occasionally, for my CTO as a Service clients. I am pretty good at C#, Java, C++, and NodeJS/TypeScript. I can also stumble around in Python and Scala.

One of the languages that I have been meaning to teach myself in Golang. I kept hearing that Go is a great language for writing distributed systems, and I certainly have written my fair share of distributed systems. I started life way back when as a C programmer, and with Golang, I feel that I have come full-circle. The nice thing about Golang is the support for writing multi-threaded applications. 

I always like to write something useful when I learn a new technology. I have been spending an increasing amount of time in Slack, and I come from the world of finance. So I figured that I could combine the two for my first application in Go.

## Before You Run The Application

To run this, you will need to create a file called `appSettings.json` and make sure that this file is in the same directory as the application. The `appSettings.json` file looks like this:

```
{
    "apiKeys": {
        "quandl": "[Your Quandl API Key]",
        "worldtrading": "[Your World Trading Data API Key]",
        "alphavantage": "[Your AlphaVantage API Key]"
    },
    "driver": "alphavantage",
    "slackSecret": "[Your app's Slack Secret]",
    "port": 5000
}
```

## Visual Studio Code launch configuration

If you are using Visual Studio Code, you will have a directory called `.vscode` that contains a file called `launch.json`.

```
{
    "version": "0.1.0",
    "configurations": [
        {
            "name": "Launch",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/application.go"
        }
    ]
}
```
