package discord

import (
	"fmt"
	"context"
	"strings"
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
	// c, err := client.GetChannel(context.Background(), disgord.Snowflake(id))
	// if err == nil {
	// 	fmt.Println("Channel: ", c.Name)
	// 	return c
	// }

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

	index := 0
	modifiedName:= fmt.Sprintf("%s-%d", channelname, index)
	for i, cha := range channels {
		fmt.Println(i)
		name := cha.Name
		splittedName := strings.Split(name, "-")
		num := splittedName[1:]
		fmt.Println("Name", splittedName)
		fmt.Println("Num ", num)
	}

	// fmt.Sprintf("%s%d", modifiedName, index)
	channel := &disgord.CreateGuildChannelParams{
		Name: modifiedName + string(index),
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
