#!/usr/bin/env python

from re import L
import wikipediaapi
import requests
import sys 
import subprocess
import json
from bs4 import BeautifulSoup
from rich import print
from rich.panel import Panel
from rich.console import Console
from rich.text import Text

def display_help():
    print("  usage: info args")
    print("  type word or phrase to search such as:")
    print("     \'info Dune\'\n")
    print("  To search text displayed:")
    print("     \'/<term>\'\n")
    exit(1)

#print sections of page
def print_sections(sections):
    index = 1
    for s in sections:
        sections_dict[index] = s.title
        print(f"{index}: [bold]{s.title}[/bold] - {s.text[0:40]}...")
        index += 1

#options
#TODO: -dev flag for embed nvim
match sys.argv[1]:
    case 'news':
        result = subprocess.run(["/home/matt/info/fetcher", sys.argv[1], sys.argv[2]], capture_output=True, text=True)
        if result.stderr:
            print(result.stderr)
            exit(1)
        news_url = result.stdout
        print(result)
        #Beatiful Soup to render html
        html = requests.get(news_url).text
        soup = BeautifulSoup(html, "html.parser")
        articles_class = soup.find_all(attrs={'class':'dcr-2yd10d'})
        print("[blue]List of articles")
        alt = 0
        num_in_art = 1
        #list out all articles in section
        for a in articles_class:
            match alt:
                case 0:
                    print(f"[green]{num_in_art}. {a.get('aria-label')}\n")
                case 1:
                    print(f"[blue]{num_in_art}. {a.get('aria-label')}\n")
            alt ^= 1
            num_in_art += 1
        #select an article 
        selected_val = int(input("Select an article\n"))
        selected_article = str(articles_class[selected_val - 1].get('href'))
        #send to fetcher
        result = subprocess.run(["/home/matt/info/fetcher", selected_article], capture_output=True, text=True)
        if result.stderr:
            print(result.stderr)
            exit(1)
        news_url = result.stdout
        #parse html
        html = requests.get(news_url).text
        soup = BeautifulSoup(html, "html.parser")
        art_title = soup.find('h1')
        art_content = soup.find_all('p')
        if art_title != None:
            print(f"\n[red]{art_title.get_text()}\n")
        else:
            print("Title unknown")
        for t in art_content:
            print(f"{t.get_text()}\n")

    #wiki 
    case 'wiki':
        #arguments for wiki summary
        page_input = " ".join(sys.argv[2:])

        if page_input == 'help' or page_input == 'h': 
            display_help()

        sections_dict = {}

        #request setup 
        url = "https://en.wikipedia.org/w/api.php"
        params = {
                "action": "opensearch",
                "namespace": "0",
                "search": page_input,
                "limit": "5",
                "format": "json"
                }
        headers = {
                "User-Agent": "infocli/1.0 (myemail@example.com)"
                }

        #these are results from input
        r = requests.get(url, params=params, headers=headers)
        data = [];
        if r.status_code == 200:
            data = r.json()
        else:
            print(r.status_code)

        if not data[1]:
            print("No valid pages found")
            exit(1)

        #title pages, as well as print available pages for help
        titles = data[1]
        print("Available pages")
        print(data[1])
        title_page = titles[0]

        #fetch the page
        wiki_wiki = wikipediaapi.Wikipedia(user_agent='infocli/1.0 (myemail@example.com)', language='en')
        page_py = wiki_wiki.page(title_page)

        if not page_py.exists():
            print("Page does not exist")
            exit(1)

        #print info
        print("[bold]Title:[/bold] %s\n" % page_py.title)
        print(Panel(page_py.summary, title="Summary"))
        p_url = page_py.fullurl
        print(f"[blue]{p_url}\n")
        while(True):
            print_sections(page_py.sections)
            sec_input = input('Input\n')
            #exit program
            if sec_input == 'q': 
                exit(1)
            #highlight keyword TODO: fuzzy?
            elif sec_input[0] == '/': 
                sec_input = sec_input[1:]
                console = Console()
                console.clear()
                text = Text(page_py.summary)
                text.highlight_words([sec_input], style = "bold blue", case_sensitive =False)
                print(Panel(text, title="Summary"))
                continue
            #get integer value of input
            try:
                sec_input = int(sec_input)
            except ValueError:
                print("  invalid input\n")
                display_help()
            selected_section = sections_dict.get(sec_input)
            #sections logic
            if not selected_section:
                print("Not valid section")
            else:
                section = page_py.section_by_title(selected_section)
                if section is not None:
                    print(Panel(section.text, title=section.title))
                else:
                    print("section is invalid")

    case _:
        display_help()
