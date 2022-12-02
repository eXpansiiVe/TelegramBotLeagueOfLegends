# About the project

It's just a Telegram bot that displays a given lol summoner's info.
Info includes summoner name, level, profile image, ranked tier, division, wins, loses and lp for SoloQ and Flex (if available).

That's how the output looks like:

![alt text](/BotShowCase.png)

# Getting Started

First of all you need a telegram bot token and a riot api key, if you don't have one of those just follow this links:\
TelegramBot: https://t.me/BotFather\
Riot ApiKey: https://developer.riotgames.com/\
Now that you have them go in the main file and put them in the corrispective constant(line 16-17):
```
const riotApiKey string = ""
const telegramBotToken string = ""
```
To make it recieve API call from telegram we have to deploy a server, so if you already have one where you can recieve calls look at the main function, put the ip inside the http.ListenAndServer function with the right port and skip the server creation part.

```
func main() {
 log.Println("Listening on port 3000")
 err := http.ListenAndServe("<IP:PORT>", http.HandlerFunc(Handler))
 if err != nil {
  log.Panic(err)
 }
}
```

If you don't have a server and you still want to deploy your server you can use 'ngrok'.
Install it and then run the following command with the port you prefer (the bot is already configured for the port 3000), example:

```
ngrok http 3000
```

Once successful, you should be able to see the public URL for your bot,
that should be something like this: \
`https://a15291fk.ngrok.io`

# Set up the Webhook

Now, all we need to do is let telegram know that our bot has to talk to this url whenever it receives any message.

So just just enter this in your terminal, changing the **URL** with the one from ngrok and put your **bot token** in the right spot.

```
curl -F "url=https://b46129ha.ngrok.io/"  https://api.telegram.org/bot<your_api_token>/setWebhook
```

N.B.: You have to use this command up here even if you already have a hosted server. 
# Usage

DONE, now you have only to run it:
```
go run ./cmd 
```
Enjoy it!

# Last thing!

Remember, that's my first project! I know it's a bit messy but if you want to improve the code feel free to do so.
