package main

import (
	"fmt"
	"os"
	"log"

	"github.com/joho/godotenv"
)

func main(){

	//api key logic
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	//g_key := os.Getenv("G_KEY")
	//search_key

	//exec, type(wiki, news, search), extra arguments
	fetch_type := os.Args[1]
	extra_args := os.Args[2:]
	var arg_string string
	for i := 0; i < len(os.Args) -2; i++ {
		arg_string = arg_string + extra_args[i]
	}

	//switch different fetch types
	switch fetch_type {
	case "news":
	case "search":
	default:
		fmt.Print("invalid option")
	}

}
