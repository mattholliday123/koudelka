package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)




//structs for guardian response
type Result struct {
    WebUrl string `json:"webUrl"`
}

type Response struct {
    Results []Result `json:"results"`
		Content Content `json:"content"`
}

type GuardianResponse struct {
    Response Response `json:"response"`
}

type Content struct {
	WebUrl string `json:"webUrl"`
}

func fetchSections(g_key string){
	g_url := "https://content.guardianapis.com/sections?&api-key=" + g_key
	var m GuardianResponse 
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
		body, body_err := io.ReadAll(response.Body)
		if body_err != nil{
			log.Fatal(err)
		}
		//decode body
		json_err := json.Unmarshal(body, &m)
		if json_err != nil{
			log.Fatal(json_err)
		}
		//return back to python script
		for _, r := range m.Response.Results {
			fmt.Println(r.WebUrl)
		}
		os.Exit(0)
}
func fetchContent(partial string, g_key string){
	g_url := "https://content.guardianapis.com" + partial + "?api-key=" + g_key
	var m GuardianResponse 
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
		body, body_err := io.ReadAll(response.Body)
		if body_err != nil{
			log.Fatal(err)
		}
		//decode body
		json_err := json.Unmarshal(body, &m)
		if json_err != nil{
			log.Fatal(json_err)
		}
		//return back to python script
			fmt.Print(m.Response.Content.WebUrl)
		//exit
		os.Exit(1)
}

func main(){

	//api key logic
	err := godotenv.Load("/home/matt/info/keys.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	g_key := os.Getenv("G_KEY")
	//search_key

	//exec, type(news, search), extra arguments
	fetch_type := os.Args[1]
	if strings.HasPrefix(fetch_type, "/"){
		fetchContent(fetch_type, g_key)
	}
	//extra_args := os.Args[2:]
	var g_url string
	var m GuardianResponse 


	if len(os.Args) == 2{
		g_url = "https://content.guardianapis.com/sections?q=us-news/us-politics&api-key=" + g_key
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
				body, body_err := io.ReadAll(response.Body)
				if body_err != nil{
					log.Fatal(err)
				}
				//decode body
				json_err := json.Unmarshal(body, &m)
				if json_err != nil{
					log.Fatal(json_err)
				}
				//return back to python script
				for _, r := range m.Response.Results {
					fmt.Println(r.WebUrl)
				}
	}
	arg_string := os.Args[2]
	//for i := 0; i < len(os.Args); i++ {
	//	arg_string += extra_args[i] + " "
	//}


		//switch different fetch types
	switch fetch_type {
	case "news":
		switch os.Args[2]{
			case "sections":
				fetchSections(g_key)
		//request
			default:
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
				body, body_err := io.ReadAll(response.Body)
				if body_err != nil{
					log.Fatal(err)
				}
				//decode body
				json_err := json.Unmarshal(body, &m)
				if json_err != nil{
					log.Fatal(json_err)
				}
				//return back to python script
				for _, r := range m.Response.Results {
					fmt.Println(r.WebUrl)
				}
			}

	case "search":
		//http.Get(url)
		//err := json.Unmarshal(response, &m)
	default:
		fmt.Print("invalid option")
	}

}
