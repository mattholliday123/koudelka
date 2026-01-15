from re import escape
from textual.app import App, ComposeResult
from textual.widgets import Static, Button, Input, Label
from textual.containers import Container, Horizontal
from textual import on

from client import send_message

class MyApp(App):
    CSS_PATH = "app.tcss"
    BINDINGS = [ 
                ("escape", "normal_mode", "Enter Normal Mode"),
                ("i", "insert_mode", "Enter Insert Mode")
                ]

    def compose(self) -> ComposeResult:
        yield Static("Koudelka", classes = "title")
        yield Horizontal(
        Input(placeholder = "search...", id="search_bar"),
        classes = "search_bar",
        )
        yield Static("--Normal--", id="indicator")

    def on_mount(self) -> None:
        search_bar = self.query_one("#search_bar", Input)
        search_bar.disabled=True
        search_bar.can_focus = False

    @on(Input.Submitted)
    def get_search(self):
        input = self.query_one(Input)
        query = input.value
        input.value = ""
        self.mount(Label(query))
        send_message(query)

    def action_normal_mode(self) -> None:
        indicator = self.query_one("#indicator", Static)
        search_bar = self.query_one("#search_bar", Input)
        indicator.update("--Normal--")
        indicator.remove_class("insert")
        indicator.add_class("normal")
        search_bar.disabled=True
        search_bar.can_focus = False
        self.set_focus(None)
    def action_insert_mode(self) -> None:
        indicator = self.query_one("#indicator", Static)
        search_bar = self.query_one("#search_bar", Input)
        indicator.update("--Insert--")
        indicator.remove_class("normal")
        indicator.add_class("insert")
        search_bar.disabled=False
        search_bar.can_focus = True 
        search_bar.focus()
        

if __name__ == "__main__": 
    app = MyApp()
    app.run()
