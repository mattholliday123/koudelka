from re import escape
from textual.app import App, ComposeResult
from textual.widgets import Static, Button, Input, Label
from textual.containers import Container, Horizontal
from textual import on

from client import send_message

class MyApp(App):
    CSS_PATH = "app.tcss"
    '''
    BINDINGS = [ 
                ("escape", "normal_mode()", "Enter Normal Mode"),
                ("i", "insert_mode()", "Enter Insert Mode")
                when in insert, can just go directly to search
                Insert normal mode to scroll and browse
                ]
    '''

    def compose(self) -> ComposeResult:
        yield Static("Koudelka", classes = "title")
        yield Horizontal(
        Input(placeholder = "search..."),
        classes = "search_bar"
        )

    @on(Input.Submitted)
    def get_search(self):
        input = self.query_one(Input)
        query = input.value
        input.value = ""
        self.mount(Label(query))
        send_message(query)


if __name__ == "__main__": 
    app = MyApp()
    app.run()
