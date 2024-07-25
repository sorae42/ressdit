# subeddit-rss

[reddit post explaining what this is](https://www.reddit.com/r/rss/comments/fvg3ed/i_built_a_better_rss_feed_for_reddit/)

**NOTE**: This fork includes **breaking changes**! You may want to see Installation for more details.

## installation

Using Go: `go build ./cmd/reddit-rss && ./reddit-rss`. 

Using Docker (recommended): 

1. Install Docker if you haven't.
2. Run `docker build .`
3. Then run `docker run -d -p<your port here>:8080 '<instance name>'` to run an subreddit-rss instance.

Server will be started at http://localhost:8080 (or whatever the port you set above).

To subscribe to a subreddit:

1. Go to a subreddit or meta feed you like example: https://www.reddit.com/r/Touhou
2. Change the domain name to the server domain: https://localhost:8080/r/Touhou
3. Subscribe to the url in your favorite feed reader.

NOTE: Please **DO NOT** append `.json` to the url path. They are deprecated in this fork.

### OAUTH

To get access to better rate limits, you can set up an oauth app on reddit and provide the client id and secret as environment variables.
Be sure to select script when asked what kind of app you are.
<https://old.reddit.com/prefs/apps/>

```
OAUTH_CLIENT_ID=your_client_id # its that id in the top left of the reddit app page
OAUTH_CLIENT_SECRET=yout_client_secret # its the secret under the id
REDDIT_USERNAME=your_reddit_username # the username of the account you created the app with
REDDIT_PASSWORD=your_reddit_password # the password of the account you created the app with
USER_AGENT="browser:name-of-app:v1.0.0 (by /u/your-reddit-username)"
```

## query params

-   `?safe=true` filter out nsfw posts
-   `?scoreLimit=100` filter out posts with less than 100 up votes
-   `?flair=Energy%20Products` only include posts that have that flair

## Docker configuration

to further configure your instance, you can set the following environment variables

### REDDIT_URL

this controls which interface you want your rss feed entries to link to (to avoid tracking and that annoying use mobile app popup). any alternative reddit interface can be provided here, ie: https://libredd.it or https://teddit.net .

it defaults to `"https://www.reddit.com"`. (yes, the new reddit interface)

### PORT

Define which port your instance is listening on. Default to `8080`