package util

import (
	"fmt"
	"context"
	"github.com/andersfylling/disgord"
)

var debug = false

// CreateCategory create a category for other channels
// client : The bot.
// id : Server's ID
// channels: list of available channels
// catergoryName: The name of the category you want to create
func CreateCategory(client *disgord.Client, id int, channels []*disgord.Channel, categoryName string) *disgord.Channel {
	if debug{
		fmt.Println("[DEBUG]: Received os:", categoryName)
	}

	for _, cha := range channels {
		if categoryName == cha.Name{
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
// channels: list of available channels
// client : The bot.
// id : Server's ID
// category: The location where the channels will the created
// channelname: The name of the channel you want to create
func CreateChannel(channels []*disgord.Channel,client *disgord.Client,id int, category *disgord.Channel, channelname string) *disgord.Channel {
	if debug{
		fmt.Println("[DEBUG]: Channel:", channelname)
	}

	index := 1
	modifiedName:= fmt.Sprintf("%s-%d", channelname, index)

	for _, cha := range channels {
		// fmt.Println(i)
		name := cha.Name
		if name == "general"{
			return nil
		}else
		
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
