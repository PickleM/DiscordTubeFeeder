package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	YOUTUBE_FEED_URL = "https://www.youtube.com/feeds/videos.xml?channel_id="
	config           = Config{}
	VideosSeen       map[string]struct{}
)

type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Text    string   `xml:",chardata"`
	Yt      string   `xml:"yt,attr"`
	Media   string   `xml:"media,attr"`
	Xmlns   string   `xml:"xmlns,attr"`
	Link    []struct {
		Text string `xml:",chardata"`
		Rel  string `xml:"rel,attr"`
		Href string `xml:"href,attr"`
	} `xml:"link"`
	ID        string `xml:"id"`
	ChannelId string `xml:"channelId"`
	Title     string `xml:"title"`
	Author    struct {
		Text string `xml:",chardata"`
		Name string `xml:"name"`
		URI  string `xml:"uri"`
	} `xml:"author"`
	Published string  `xml:"published"`
	Entry     []Entry `xml:"entry"`
}

type Entry struct {
	Text      string `xml:",chardata"`
	ID        string `xml:"id"`
	VideoId   string `xml:"videoId"`
	ChannelId string `xml:"channelId"`
	Title     string `xml:"title"`
	Link      struct {
		Text string `xml:",chardata"`
		Rel  string `xml:"rel,attr"`
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Author struct {
		Text string `xml:",chardata"`
		Name string `xml:"name"`
		URI  string `xml:"uri"`
	} `xml:"author"`
	Published string `xml:"published"`
	Updated   string `xml:"updated"`
	Group     struct {
		Text    string `xml:",chardata"`
		Title   string `xml:"title"`
		Content struct {
			Text   string `xml:",chardata"`
			URL    string `xml:"url,attr"`
			Type   string `xml:"type,attr"`
			Width  string `xml:"width,attr"`
			Height string `xml:"height,attr"`
		} `xml:"content"`
		Thumbnail struct {
			Text   string `xml:",chardata"`
			URL    string `xml:"url,attr"`
			Width  string `xml:"width,attr"`
			Height string `xml:"height,attr"`
		} `xml:"thumbnail"`
		Description string `xml:"description"`
		Community   struct {
			Text       string `xml:",chardata"`
			StarRating struct {
				Text    string `xml:",chardata"`
				Count   string `xml:"count,attr"`
				Average string `xml:"average,attr"`
				Min     string `xml:"min,attr"`
				Max     string `xml:"max,attr"`
			} `xml:"starRating"`
			Statistics struct {
				Text  string `xml:",chardata"`
				Views string `xml:"views,attr"`
			} `xml:"statistics"`
		} `xml:"community"`
	} `xml:"group"`
}

type Config struct {
	DiscordToken     string
	DiscordChannelID string
	YoutubeChannelID string
}

func main() {
	VideosSeen = make(map[string]struct{})

	config = Config{
		DiscordToken:     os.Getenv("DISCORD_TOKEN"),
		DiscordChannelID: os.Getenv("DISCORD_CHANNEL_ID"),
		YoutubeChannelID: os.Getenv("YOUTUBE_CHANNEL_ID"),
	}
	discord, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		log.Fatal(err)
	}

	err = discord.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer discord.Close()

	if len(VideosSeen) == 0 {
		log.Println("Treating all current videos as being seen")

		feed, err := GetYoutubeFeed(config.YoutubeChannelID)
		if err != nil {
			log.Fatal(err)
		}

		for _, video := range feed.Entry {
			VideosSeen[video.ID] = struct{}{}
		}
	}

	for {
		log.Println("Checking for new videos")

		video, found, err := CheckForNewVideos()
		if err != nil {
			log.Printf("Error checking for new video: %s\n", err)
		}

		if found {
			log.Printf("Found a new video: %s\n", video.Title)

			discord.ChannelMessageSend(config.DiscordChannelID, video.Link.Href)
		}
		time.Sleep(10 * time.Second)
	}
}

// CheckForNewVideos fetches the XML feed from youtube and compares the entries to VideosSeen
func CheckForNewVideos() (Entry, bool, error) {
	feed, err := GetYoutubeFeed(config.YoutubeChannelID)
	if err != nil {
		return Entry{}, false, err
	}

	// Loop over all the youtube videos
	for _, video := range feed.Entry {
		_, seen := VideosSeen[video.ID]
		if !seen {
			VideosSeen[video.ID] = struct{}{}
			return video, true, nil
		}
	}

	return Entry{}, false, nil
}

// GetYoutubeFeed fetches the XML feed for the given channelID and returns a Feed{}
func GetYoutubeFeed(channelID string) (Feed, error) {
	url := YOUTUBE_FEED_URL + channelID

	// Fetch the XML
	res, err := http.Get(url)
	if err != nil {
		return Feed{}, err
	}

	// Copy the http body out of the http request
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Feed{}, err
	}

	// Unmarshal the XML
	feed := Feed{}
	err = xml.Unmarshal(content, &feed)
	if err != nil {
		return Feed{}, err
	}

	return feed, nil
}
