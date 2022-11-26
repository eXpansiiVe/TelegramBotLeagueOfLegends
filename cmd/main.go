package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/eXpansiiVe/LoLBot/pkg/schemes"
)

const apiKey string = "RGAPI-fd2cdcde-e835-44da-8648-4c691175c919"

// Used to get a response from an API
func getRequestData(link, errorMessage string) io.ReadCloser {
	rawResponseData, err := http.Get(link)
	fmt.Printf("Status code: %v\n", rawResponseData.StatusCode)
	if err != nil {
		fmt.Println(errorMessage, err)
		return nil
	}

	return rawResponseData.Body
}

// A simple reader use to read the ResponseDate passed by parameter
func filterRequestData(rawResponseData io.ReadCloser, errorMessage string) []byte {
	responseData, err := io.ReadAll(rawResponseData)
	if err != nil {
		fmt.Println(errorMessage, err)
		return nil
	}

	return responseData
}

// Get the summoner id through API
func getAccountID(nickname string) []byte {
	accountLink := "https://euw1.api.riotgames.com/lol/summoner/v4/summoners/by-name/" + url.QueryEscape(nickname) + "?api_key=" + url.QueryEscape(apiKey)

	rawResponseAccountData := getRequestData(accountLink, "Got an error retrieving the Account ID api")

	responseAccountData := filterRequestData(rawResponseAccountData, "Got an error reading the account ID body")
	defer func(rawResponseAccountData io.ReadCloser) {
		err := rawResponseAccountData.Close()
		if err != nil {
			fmt.Println("Got an error closing the response account data")
		}
	}(rawResponseAccountData)
	return responseAccountData
}

/*
func getPlayerInfo(m map[string]string, s schemes.LoLAccount, numRank int) map[string]string {
	switch numRank {
	case 0:
		m["Nickname"] = fmt.Sprintf("%v\n", s[0].SummonerName)
		m["Lp"] = "0"
		m["Wins"] = "0"
		m["Loses"] = "0"
	case 1:
		m["Nickname"] = fmt.Sprintf("%v\n", s[0].SummonerName)
		m["Lp"] = fmt.Sprintf("%v\n", s[0].LeaguePoints)
		m["Wins"] = fmt.Sprintf("%v\n", s[0].Wins)
		m["Loses"] = fmt.Sprintf("%v\n", s[0].Losses)
	case 2:
		m["Nickname"] = fmt.Sprintf("%v\n", s[0].SummonerName)
		m["Lp"] = fmt.Sprintf("%v\n", s[0].LeaguePoints)
		m["Wins"] = fmt.Sprintf("%v\n", s[0].Wins)
		m["Loses"] = fmt.Sprintf("%v\n", s[0].Losses)
		m["Lp2"] = fmt.Sprintf("%v\n", s[1].LeaguePoints)
		m["Wins2"] = fmt.Sprintf("%v\n", s[1].Wins)
		m["Loses2"] = fmt.Sprintf("%v\n", s[1].Losses)
	}
	return m
}*/

// Get the summoner Data through API
func getSummonerData(id string) []byte {

	playerDataLink := "https://euw1.api.riotgames.com/lol/league/v4/entries/by-summoner/" + url.QueryEscape(id) + "?api_key=" + url.QueryEscape(apiKey)

	rawResponsePlayerData := getRequestData(playerDataLink, "Got an error retrieving the summoner data")

	responsePlayerData := filterRequestData(rawResponsePlayerData, "Got an error reading the summoner Data body")

	fmt.Println("response", string(responsePlayerData))
	// TODO: check for a response of 403

	defer func(rawResponsePlayerData io.ReadCloser) {
		err := rawResponsePlayerData.Close()
		if err != nil {
			fmt.Println("Got an error closing the response account data")
		}
	}(rawResponsePlayerData)
	return responsePlayerData
}

// Parse the data to check if there's a rank and return it
func filterRankData(s schemes.LoLAccount, m map[string]string) error {
	var schemeLen = len(s)
	var noRankError error

	if schemeLen > 0 {
		m["Nickname"] = fmt.Sprintf("%s\n", s[0].SummonerName)
		for i := 0; i < schemeLen; i++ {
			if s[i].QueueType == "RANKED_SOLO_5x5" {
				m["RankSoloQ"] = fmt.Sprintf("%s %s\n", s[i].Tier, s[i].Rank)
				m["LpQ"] = fmt.Sprintf("%v", s[i].LeaguePoints)
				m["WinsQ"] = fmt.Sprintf("%v\n", s[i].Wins)
				m["LosesQ"] = fmt.Sprintf("%v\n", s[i].Losses)
			} else {
				m["RankFlex"] = fmt.Sprintf("%s %s\n", s[i].Tier, s[i].Rank)
				m["LpFlex"] = fmt.Sprintf("%v", s[i].LeaguePoints)
				m["WinsFlex"] = fmt.Sprintf("%v\n", s[i].Wins)
				m["LosesFlex"] = fmt.Sprintf("%v\n", s[i].Losses)
			}
		}
		return nil
	}
	noRankError = errors.New("rank not found")
	return noRankError

}

// Create a formatted text message for telegram
func messageTextFormatter(m map[string]string) string {
	//numId := strings.Replace(m["imgId"], "\n", "", 1)

	//imgLink := "http://ddragon.leagueoflegends.com/cdn/12.16.1/img/profileicon/" + numId + ".png"
	imgLink := "http://ddragon.leagueoflegends.com/cdn/12.16.1/img/profileicon/7.png"

	formattedText := fmt.Sprintf("[](%v)\n**Nome:** %v**Livello:** %v\n\n"+
		"**FlexQ**\n **Lega:** %v **Vittorie:** %v **Sconfitte:** %v **Lp:** %v\n"+
		"**SoloQ**\n\n **Lega:** %v **Vittorie:** %v **Sconfitte:** %v **Lp:** %v",
		imgLink, m["Nickname"], m["Level"], m["RankFlex"], m["WinsFlex"], m["LosesFlex"], m["LpFlex"],
		m["RankSoloQ"], m["WinsQ"], m["LosesQ"], m["LpQ"])

	return formattedText
}

// Get telegram data through API
func getTelegramApi() []byte {
	rawResponseTelegramData := getRequestData("https://api.telegram.org/bot5683492318:AAFW8Yt40ggMfd7eP5p-Ea1pzao2G_oAgsg/getUpdates?offset=-1", "Got an error retrieving telegram response")
	responseTelegramData := filterRequestData(rawResponseTelegramData, "Got an error while reading the body")

	defer func(rawResponseTelegramData io.ReadCloser) {
		err := rawResponseTelegramData.Close()
		if err != nil {
			fmt.Println("Got an error closing the response account data")
		}
	}(rawResponseTelegramData)
	return responseTelegramData
}

// Send response to telegram through API and return a response
func sendHttpMessage(chatId int64, messageId int, message string) ([]byte, error) {
	chatIdString := fmt.Sprintf("%d", chatId)
	messageIdString := fmt.Sprintf("%d", messageId)
	formattedUrl := "https://api.telegram.org/bot5683492318:AAFW8Yt40ggMfd7eP5p-Ea1pzao2G_oAgsg/sendMessage?chat_id=" + url.QueryEscape(chatIdString) + "&reply_to_message_id=" + url.QueryEscape(messageIdString) + "&parse_mode=markdown&text=" + url.QueryEscape(message)

	rawResponseData, err := http.Get(formattedUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(rawResponseData.Body)
	responseData, err := io.ReadAll(rawResponseData.Body)
	if err != nil {
		return nil, err
	}
	return responseData, nil
}

func main() {

	var updateID int

	for {
		// Instantiate every schemes
		var schemeTg schemes.ApiTelegram
		var schemeAccount schemes.AccountRiot
		var schemeLoLData schemes.LoLAccount
		var schemeTgMessageResponse schemes.ApiTelegramMessage
		var schemePlayerNotFound schemes.PlayerNotFound

		playerInfo := make(map[string]string)

		// Get telegram response data
		rawTelegramResponseData := getTelegramApi()
		err := json.Unmarshal(rawTelegramResponseData, &schemeTg)
		if err != nil {
			fmt.Println("Got error unmarshal raw telegram response data")
		}
		// Sleep the app to not spam too much requests
		fmt.Println("Sleeping")
		time.Sleep(5 * time.Second)

		// Check to not cycle to the same message
		if updateID == schemeTg.Result[0].Message.MessageID {
			fmt.Println("Inside if")
			continue
		}
		// Get the id of the riot account
		responseAccountData := getAccountID(schemeTg.Result[0].Message.Text)

		// Check if the player exist, if not send a message, assign to updateID the new value and goes to new cycle
		err = json.Unmarshal(responseAccountData, &schemePlayerNotFound)
		if err != nil {
			fmt.Println("Got an error unmarshal response account data to schemePlayerNotFound")
		}
		if schemePlayerNotFound.Status.StatusCode == 404 {
			message := "Player not found!"
			responseMessage, err := sendHttpMessage(schemeTg.Result[0].Message.Chat.ID, schemeTg.Result[0].Message.MessageID, message)
			if err != nil {
				log.Println(err)
			}

			err = json.Unmarshal(responseMessage, &schemeTgMessageResponse)
			if err != nil {
				fmt.Println("Got error unmarshal response message")
			}
			fmt.Println("Response from telegram message ", schemeTgMessageResponse.Ok)
			//Assign the messageId to the var updateID to not cycle on the same message

			updateID = schemeTg.Result[0].Message.MessageID
			continue
		}

		err = json.Unmarshal(responseAccountData, &schemeAccount)
		if err != nil {
			fmt.Println("Got error unmarshal response account data")
		}

		// Write summoner level to the map
		playerInfo["Level"] = fmt.Sprintf("%v", schemeAccount.SummonerLevel)
		playerInfo["imgId"] = fmt.Sprintf("%v\n", schemeAccount.ProfileIconID)

		// Get the summoner data response
		rankData := getSummonerData(schemeAccount.ID)
		err = json.Unmarshal(rankData, &schemeLoLData)
		if err != nil {
			fmt.Println("Got error unmarshal rank data")
		}

		// Check if there's a schemeLoLData.QueueType and if there's any take the actual rank
		err = filterRankData(schemeLoLData, playerInfo)

		if err != nil {
			log.Println(err)
		}

		message := messageTextFormatter(playerInfo)

		// Send a message to the message sender on telegram with the results
		responseMessage, err := sendHttpMessage(schemeTg.Result[0].Message.Chat.ID, schemeTg.Result[0].Message.MessageID, message)
		if err != nil {
			log.Println(err)
		}

		err = json.Unmarshal(responseMessage, &schemeTgMessageResponse)
		if err != nil {
			fmt.Println("Got error unmarshal response message")
		}
		fmt.Println("Response from telegram message ", schemeTgMessageResponse.Ok)
		//Assign the messageId to the var updateID to not cycle on the same message

		updateID = schemeTg.Result[0].Message.MessageID

	}
}
