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
func CreateCategory(client *disgord.Client, id int, channels []*disgord.Channel, categoryName string) *disgord.Channel {
	if debug{
		fmt.Println("[DEBUG]: Received os:", categoryName)
	}

	for _, cha := range channels {
		// fmt.Println("Channel: ", cha.Name)
		if categoryName == cha.Name{
			fmt.Println("Category already exist")
			return cha
		}
	}

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
func CreateChannel(channels []*disgord.Channel,client *disgord.Client,id int, category *disgord.Channel, channelname string) *disgord.Channel {
	if debug{
		fmt.Println("[DEBUG]: Channel:", channelname)
	}

	index := 1
	modifiedName:= fmt.Sprintf("%s-%d", channelname, index)
	for _, cha := range channels {
		// fmt.Println(i)
		name := cha.Name
		// if the name already exist, increase the team num
		if name == modifiedName{
			//increase index
			index++
			// Update the old modifiedName
			modifiedName = fmt.Sprintf("%s-%d", channelname, index)
		}
	}

	// fmt.Sprintf("%s%d", modifiedName, index)
	channel := &disgord.CreateGuildChannelParams{
		Name: modifiedName,
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
