package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
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

// Get the summoner Data through API
func getSummonerData(id string) []byte {

	playerDataLink := "https://euw1.api.riotgames.com/lol/league/v4/entries/by-summoner/" + url.QueryEscape(id) + "?api_key=" + url.QueryEscape(apiKey)

	rawResponsePlayerData := getRequestData(playerDataLink, "Got an error retrieving the summoner data")

	responsePlayerData := filterRequestData(rawResponsePlayerData, "Got an error reading the summoner Data body")

	fmt.Println("response from getting summoner data", string(responsePlayerData))

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
		m["Nickname"] = fmt.Sprintf("%s", s[0].SummonerName)
		for i := 0; i < schemeLen; i++ {
			if s[i].QueueType == "RANKED_SOLO_5x5" {
				m["RankSoloQ"] = fmt.Sprintf("%s %s", s[i].Tier, s[i].Rank)
				m["LpQ"] = fmt.Sprintf("%v", s[i].LeaguePoints)
				m["WinsQ"] = fmt.Sprintf("%v", s[i].Wins)
				m["LosesQ"] = fmt.Sprintf("%v", s[i].Losses)
			} else {
				m["RankFlex"] = fmt.Sprintf("%s %s", s[i].Tier, s[i].Rank)
				m["LpFlex"] = fmt.Sprintf("%v", s[i].LeaguePoints)
				m["WinsFlex"] = fmt.Sprintf("%v", s[i].Wins)
				m["LosesFlex"] = fmt.Sprintf("%v", s[i].Losses)
			}
		}
		return nil
	}
	noRankError = errors.New("rank not found")
	return noRankError

}

// Create a formatted text message for telegram
func messageTextFormatter(m map[string]string, imgId string, s schemes.LoLAccount) (string, string) {
	imgLink := "https://ddragon.leagueoflegends.com/cdn/12.16.1/img/profileicon/" + url.QueryEscape(imgId) + ".png"
	schemeLen := len(s)
	formattedText := fmt.Sprintf("<b>Nome:</b> %v\n", m["Nickname"])
	formattedText += fmt.Sprintf("<b>Livello:</b> %v\n\n", m["Level"])

	if schemeLen == 1 {
		if s[0].QueueType == "RANKED_SOLO_5x5" {
			formattedText += fmt.Sprintf("<b>SoloQ</b> \n")
			formattedText += fmt.Sprintf("<b>Lega:</b> %v\n", m["RankSoloQ"])
			formattedText += fmt.Sprintf("<b>Vittorie:</b> %v\n", m["WinsQ"])
			formattedText += fmt.Sprintf("<b>Sconfitte:</b> %v\n", m["LosesQ"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", m["LpQ"])
		} else {
			formattedText += fmt.Sprintf("<b>Flex</b>\n")
			formattedText += fmt.Sprintf("<b>Lega:</b> %v\n", m["RankFlex"])
			formattedText += fmt.Sprintf("<b>Vittorie:</b> %v\n", m["WinsFlex"])
			formattedText += fmt.Sprintf("<b>Sconfitte:</b> %v\n", m["LosesFlex"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", m["LpFlex"])
		}
	} else {
		if s[0].QueueType == "RANKED_SOLO_5x5" || s[1].QueueType == "RANKED_SOLO_5x5" {
			formattedText += fmt.Sprintf("<b>SoloQ</b> \n")
			formattedText += fmt.Sprintf("<b>Lega:</b> %v\n", m["RankSoloQ"])
			formattedText += fmt.Sprintf("<b>Vittorie:</b> %v\n", m["WinsQ"])
			formattedText += fmt.Sprintf("<b>Sconfitte:</b> %v\n", m["LosesQ"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", m["LpQ"])
		}
		if s[0].QueueType == "RANKED_FLEX_SR" || s[1].QueueType == "RANKED_FLEX_SR" {
			formattedText += fmt.Sprintf("<b>Flex</b>\n")
			formattedText += fmt.Sprintf("<b>Lega:</b> %v\n", m["RankFlex"])
			formattedText += fmt.Sprintf("<b>Vittorie:</b> %v\n", m["WinsFlex"])
			formattedText += fmt.Sprintf("<b>Sconfitte:</b> %v\n", m["LosesFlex"])
			formattedText += fmt.Sprintf("<b>Lp:</b> %v\n\n", m["LpFlex"])
		}
	}
	return formattedText, imgLink
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

// Send photo message to telegram through API and return a response
func sendPhotoMessage(chatId int64, messageId int, message string, link string) ([]byte, error) {
	chatIdString := fmt.Sprintf("%d", chatId)
	messageIdString := fmt.Sprintf("%d", messageId)
	fmt.Printf("%v\n%v\n%v\n%v\n", chatIdString, messageIdString, message, link)
	formattedUrl := "https://api.telegram.org/bot5683492318:AAFW8Yt40ggMfd7eP5p-Ea1pzao2G_oAgsg/sendPhoto?chat_id=" + url.QueryEscape(chatIdString) + "&reply_to_message_id=" + url.QueryEscape(messageIdString) + "&photo=" + url.QueryEscape(link) + "&caption=" + url.QueryEscape(message) + "&parse_mode=html"

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

// Send message to telegram through API and return a response
func sendMessage(chatId int64, messageId int, message string) ([]byte, error) {
	chatIdString := fmt.Sprintf("%d", chatId)
	messageIdString := fmt.Sprintf("%d", messageId)
	formattedUrl := "https://api.telegram.org/bot5683492318:AAFW8Yt40ggMfd7eP5p-Ea1pzao2G_oAgsg/sendMessage?chat_id=" + url.QueryEscape(chatIdString) + "&reply_to_message_id=" + url.QueryEscape(messageIdString) + "&text=" + url.QueryEscape(message)

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
		if !strings.HasPrefix(schemeTg.Result[0].Message.Text, "/euw") {
			continue
		}
		schemeTg.Result[0].Message.Text = strings.ReplaceAll(schemeTg.Result[0].Message.Text, "/euw", "")
		// Get the id of the riot account
		responseAccountData := getAccountID(schemeTg.Result[0].Message.Text)

		// Check if the player exist, if not send a message, assign to updateID the new value and goes to new cycle
		err = json.Unmarshal(responseAccountData, &schemePlayerNotFound)
		if err != nil {
			fmt.Println("Got an error unmarshal response account data to schemePlayerNotFound")
		}
		if schemePlayerNotFound.Status.StatusCode == 404 {
			message := fmt.Sprintf("L'username %s non è valido!", schemeTg.Result[0].Message.Text)
			responseMessage, err := sendMessage(schemeTg.Result[0].Message.Chat.ID, schemeTg.Result[0].Message.MessageID, message)
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
		// Write imgId to var
		imgId := fmt.Sprintf("%v", schemeAccount.ProfileIconID)

		// Get the summoner data response
		rankData := getSummonerData(schemeAccount.ID)
		err = json.Unmarshal(rankData, &schemeLoLData)
		if err != nil {
			fmt.Println("Got error unmarshal rank data")
		}

		// Check if there's a schemeLoLData.QueueType and if there's any take the actual rank
		err = filterRankData(schemeLoLData, playerInfo)

		// If player rank not found
		if err != nil {
			log.Println(err)
			message := fmt.Sprintf("Il rank del player %s non è stato trovato!", schemeTg.Result[0].Message.Text)
			// Send a message to the message sender on telegram with the results
			responseMessage, err := sendMessage(schemeTg.Result[0].Message.Chat.ID, schemeTg.Result[0].Message.MessageID, message)

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

		// format the message with the playerInfo data
		message, imgLink := messageTextFormatter(playerInfo, imgId, schemeLoLData)

		// Send a message to the message sender on telegram with the results
		responseMessage, err := sendPhotoMessage(schemeTg.Result[0].Message.Chat.ID, schemeTg.Result[0].Message.MessageID, message, imgLink)
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
