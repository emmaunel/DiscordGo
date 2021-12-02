<h1 align="center"> DiscordGo </h1> 
<p align="center">
  <a>
    <img src="./img/logo.png" width="400">
  </a>
</p>



![Version](https://img.shields.io/badge/Version-2.0-brightgreen)
![Language](https://img.shields.io/badge/Language-Go-blue.svg?longCache=true&style=flat-square) 


Discord C2 for Redteam engagement....Need a better name.
If you can think of one, please tell me. :)

# Why I made this

During Blue-Red Team competition, I needed an easy and fast way to keep connected and a way for mutiple redteamer to run commands, hence DiscordGo.
Since Discord is getting popular, why not use the platorm as a c2.
That's what this project is about.

# Installation

To use DiscordGo, you need to create a Discord bot and a Discord server. After that, invite the bot to your server.

Click [here](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server-) to learn how to create a server and [here](https://discordjs.guide/preparations/setting-up-a-bot-application.html#creating-your-bot) to create a bot. And finally, learn to invite the bot to your server with [this.](https://discordjs.guide/preparations/adding-your-bot-to-servers.html#bot-invite-links)

When creating the bot, you need it give it some permission. For testing, I gave the bot full `administrative` permission. But the required permission are as follow:

* Send Messages
* Read Messages
* Attach Files
* Manage Server

# Usage

Edit this file `pkg/util/constants/variables.go` with your token and ID. Or create it if not there

The bot token can be found on discord developer dashboard where you created the bot. To get your server ID, go to your server setting and click on `widget`. On the right pane, you see the your ID.

An example configuration file looks like this:
```
var ServerID = "XXXXXXXXXXXX"
var BotToken = "XXXXXXXXXXX"
```

After that is done, all you have to do is run `make`. That will create 5 binaries.

```
* d2Server --> MacOS Server binary(tested on Big Sur)
* lind2Server --> Linux Server binary(should work on most linux distro)
* linux-agent
* windows-agent.exe
* macos-agent
```

# Feature

* Cross-platform
* Organozer(talk about and intergration to pwnboard)


# WIP (Work in Progress)

- [x] Cross-platform
- [ ] File upload
- [ ] File download
- [ ] Agent grouping(by hostname like web hosts and so on, slash command)
- [ ] Group commands
- [ ] Integrate with pwndrop



# Screenshots
<p align="center">
  <a>
    <img src="./img/example-team.png" width="860">
  </a>
</p>
<p align="center">
  <a>
    <img src="./img/example-command.png" width="860">
  </a>
</p>


# Disclamers
The author is in no way responsible for any illegal use of this software. It is provided purely as an educational proof of concept. I am also not responsible for any damages or mishaps that may happen in the course of using this software. Use at your own risk.

Every message on discord are saved on Discord's server, so be careful and not upload any sensitive or confidential documents.

# Used Libraries
* [discordgo](https://github.com/bwmarrin/discordgo)


Inspired by [SierrOne](https://github.com/berkgoksel/SierraOne)