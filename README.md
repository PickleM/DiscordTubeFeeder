# DiscordTubeFeeder
This is some shitty discord bot that looks at a youtube xml feed and publishes new videos. This is untested as my YouTube account is fucked and doesn't work with the feed for some reason.

## Usage
It requires three environment variables be set:

    DISCORD_TOKEN
    DISCORD_CHANNEL_ID
    YOUTUBE_CHANNEL_ID

Running:

    DISCORD_TOKEN=FOO DISCORD_CHANNEL_ID=BAR YOUTUBE_CHANNEL_ID=BAZ go run main.go