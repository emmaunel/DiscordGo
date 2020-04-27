package discord

import (
	"fmt"
	"context"
	"github.com/andersfylling/disgord"
)

var debug = true

// CreateCategory sends a message to the given channel.
// channelID : The ID of a Channel.
// data      : The message struct to send.
func CreateCategory(client *disgord.Client, id int, categoryName string) *disgord.Channel {
	if debug{
		fmt.Println("[DEBUG]: Received os:", categoryName)
	}

	//TODO; Check if category already exist
	category := &disgord.CreateGuildChannelParams{
		Name: categoryName,
		Type: 4,
	}

	channel, _ := client.CreateGuildChannel(context.Background(), disgord.Snowflake(id), "hello", category)
	
	return channel
}

// CreateChannel sends a message to the given channel.
// channelID : The ID of a Channel.
// data      : The message struct to send.
func CreateChannel(client *disgord.Client,id int, category *disgord.Channel, channelname string) *disgord.Channel {
	if debug{
		fmt.Println("[DEBUG]: Channel:", channelname)
	}
	channel := &disgord.CreateGuildChannelParams{
		Name: channelname,
		ParentID: category.ID,
	}
	
	//fmt.Println("ParentID ", category.ID)
	subchan, _ := client.CreateGuildChannel(context.Background(), disgord.Snowflake(id), "hello", channel)

	return subchan
}

// DeleteChannel will delete a channel given an id
// client: 
// id: Channel ID
func DeleteChannel(client *disgord.Client, id int){
	client.DeleteChannel(context.Background(), disgord.Snowflake(id))
}
