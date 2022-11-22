package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const apiKey string = "RGAPI-e719f2a1-8946-46fe-82f8-9c9a56c6eee8"

func getAccountID(nickname string) (string, error) {
	var schemeAccount AccountRiot
	accountLink := "https://euw1.api.riotgames.com/lol/summoner/v4/summoners/by-name/" + url.QueryEscape(nickname) + "?api_key=" + url.QueryEscape(apiKey)

	response, err := http.Get(accountLink)
	if err != nil {
		fmt.Println("Got an error retrieving the Account ID api", err)
		return "", err
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Got an error reading the account ID body", err)
		return "", err
	}

	json.Unmarshal(responseData, &schemeAccount)

	return schemeAccount.ID, nil
}

func getRank(id string) (string, error) {
	var schemeLoLData LoLAccount
	rankLink := "https://euw1.api.riotgames.com/lol/league/v4/entries/by-summoner/" + url.QueryEscape(id) + "?api_key=" + url.QueryEscape(apiKey)

	response, err := http.Get(rankLink)
	if err != nil {
		fmt.Println("Got an error retrieving the rank ", err)
		return "", err
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Got an error reading the lolData body", err)
		return "", err
	}

	json.Unmarshal(responseData, &schemeLoLData)

	playerRank := fmt.Sprintf("%s %s", schemeLoLData[0].Tier, schemeLoLData[0].Rank)

	return playerRank, nil

}
func main() {
	var schemeTg ClassTelegram
	var updateID int
	for {
		response, err := http.Get("https://api.telegram.org/bot5683492318:AAFW8Yt40ggMfd7eP5p-Ea1pzao2G_oAgsg/getUpdates?offset=-1")
		if err != nil {
			fmt.Println("Got an error: ", err)
			return
		}
		defer response.Body.Close()
		responseData, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Got an error while reading the body: ", err)
			return
		}
		json.Unmarshal(responseData, &schemeTg)

		if updateID != schemeTg.Result[0].Message.MessageID {
			accountID, err := getAccountID(schemeTg.Result[0].Message.Text)
			if err != nil {
				fmt.Println("Error returning the account id: ", err)
				return
			}
			fmt.Println(accountID)
			rank, err := getRank(accountID)
			if err != nil {
				fmt.Println("Error returning the rank", err)
				return
			}
			fmt.Println(rank)
			fmt.Println(schemeTg.Result[0].Message.Text)
			updateID = schemeTg.Result[0].Message.MessageID
		}

		fmt.Println("Sleeping")
		time.Sleep(5 * time.Second)
	}
}
