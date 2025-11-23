package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)


type g_response struct {
	webUrl string `json:"webUrl"`
}

func main(){

	//api key logic
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	g_key := os.Getenv("G_KEY")
	//search_key

	//exec, type(wiki, news, search), extra arguments
	fetch_type := os.Args[1]
	extra_args := os.Args[2:]
	var arg_string string
	for i := 0; i < len(os.Args) -2; i++ {
		arg_string = arg_string + extra_args[i]
	}

	var g_url string
	var m g_response

	//switch different fetch types
	switch fetch_type {
	case "news":
		//default
		//sections

		//request
		g_url = "https://content.guardianapis.com/sections?q=" + arg_string + "&api-key=" + g_key
		response, err := http.Get(g_url)
		if err != nil {
			log.Fatal(err)
		}
		//json body
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			fmt.Println("status code not ok")
			os.Exit(1)
		}
		body, err := io.ReadAll(response.Body)
		if err != nil{
			log.Fatal(err)
		}
		//decode body
		json_err := json.Unmarshal(body, &m)
		if json_err != nil{
			log.Fatal(json_err)
		}
		//return back to python script
		fmt.Print(m.webUrl)

	case "search":
		//http.Get(url)
		//err := json.Unmarshal(response, &m)
	default:
		fmt.Print("invalid option")
	}

}
