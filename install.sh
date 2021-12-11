#!/bin/sh

# TODO
# Installation script for discordC2.
echo "Beginning DiscordGo C2 installation."

printf "Please enter your server ID: "
read server
printf "Please enter your discord app token: "
read token

# Then run make
make clean && make