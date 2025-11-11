#!/usr/bin/env python

import wikipediaapi
import requests
import sys 
from rich import print
from rich.panel import Panel

sections_dict = {}

#print sections of page
def print_sections(sections):
    index = 1
    for s in sections:
        sections_dict[index] = s.title
        print(f"{index}: [bold]{s.title}[/bold] - {s.text[0:40]}...")
        index += 1

page_input = " ".join(sys.argv[1:])


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

titles = data[1]
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
print_sections(page_py.sections)
sec_input = input('input selection\n')

#wait for input to display sections
if sec_input == 'q': 
    exit(1)
sec_input = int(sec_input)
selected_section = sections_dict.get(sec_input)
if not selected_section:
    print("Not valid section")
else:
    section = page_py.section_by_title(selected_section)
    if section is not None:
        print(Panel(section.text, title=section.title))
    else:
        print("section is invalid")

