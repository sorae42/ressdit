# Ressdit

A fork of trashhalo/reddit-rss with improved (with opinion) functionality.

[See Introduction](https://www.reddit.com/r/rss/comments/fvg3ed/i_built_a_better_rss_feed_for_reddit/)

**NOTE**: This fork may includes **breaking changes**! You may want to see Installation for more details.

## Installation

Using Go: `go build ./cmd/ressdit && ./ressdit`. 

Using Docker (recommended): 

1. Install Docker if you haven't.
2. Run `docker build .`
3. Then run `docker run -d -p<your port here>:8080 '<instance name>'` to run an ressdit instance.

Server will be started at http://localhost:8080 (or whatever the port you set above).

To subscribe to a subreddit:

1. Go to a subreddit or meta feed you like example: https://www.reddit.com/r/Touhou
2. Change the domain name to the server domain: https://localhost:8080/r/Touhou
3. Subscribe to the url in your favorite feed reader.

NOTE: Please **DO NOT** append `.json` to the url path. They are deprecated in this fork.

### OAUTH

To get access to better rate limits and be able to see post in private subreddits you have joined, you must set up an oauth app on reddit and provide infomations as environment variables.

[Create an app here (open old reddit)](https://old.reddit.com/prefs/apps/)

Be sure to select **"script"** when asked what kind of app you are.

Use your instance URL as redirect URI.

You can create an `.env` file in the same directory as the executable.

```
OAUTH_CLIENT_ID=your_client_id # its that id in the top left of the reddit app page
OAUTH_CLIENT_SECRET=yout_client_secret # its the secret under the id
REDDIT_USERNAME=your_reddit_username # the username of the account you created the app with
REDDIT_PASSWORD=your_reddit_password # the password of the account you created the app with
USER_AGENT="browser:name-of-app:v1.0.0 (by /u/your-reddit-username)"
```

### Query Parameters

-   `?safe=true` filter out nsfw posts
-   `?scoreLimit=100` filter out posts with less than 100 up votes
-   `?flair=Energy%20Products` only include posts that have that flair

## Dockerfile configuration

### REDDIT_URL

This controls which interface you want your rss feed entries to link to (to avoid tracking and that annoying use mobile app popup). any alternative reddit interface can be provided here, ie: https://libredd.it or https://teddit.net .

Currently default to `"https://www.reddit.com"`. (yes, the new reddit interface)

### PORT

Define which port your instance is listening on. Default to `8080`.

You should leave this as default and use Docker port mapping instead.

## Credits

reddit-rss built by @trashhalo. [See original contributors](https://github.com/trashhalo/reddit-rss/graphs/contributors).

Forked and maintained by @sorae42.