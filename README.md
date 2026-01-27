

Demo of the browser



https://github.com/user-attachments/assets/52c99a01-8a81-4283-817b-e00e5f5558e7



This project started off as a CLI tool that fetches the wikipedia summary and sections, if desired, straight to the terminal as well as fetches news from The Guardian.
As I continued development it eventually turned into a minimal TUI browser. 

Demo of CLI tool:


https://github.com/user-attachments/assets/5a663055-f3c0-492c-becc-7a80e4af1ae7



Supports subprhase input. Example: 'Dune 19' will give the page for Dune(1984 film), or 'Dune novel' will give page for Dune(novel)


Use following command for more info 
```
info h
```
OR
```
info help
```
---------------------------------------------------------------------------------------------



NOTE: Must have these dependencies installed

```
pip install rich
```

```
pip install wikipedia-api
```

---------------------------------------------------------------------------------------------
Build instructions only for the CLI tool:


To work as system wide command:
install code
create virtual enviroment in project directory:
```
python -m venv my-venv
```
create bash script named info (or whatever you like) as 
```
#!/bin/bash
DIR="/home/username/info" #or wherever the code is located
source "$DIR/my-venv/bin/activate"
python "$DIR/info.py" "$@"
```
move script (if you named it differently, substitute info with your chosen file name)
```
sudo mv info /usr/local/bin
```

---------------------------------------------------------------------------------------------

NOTES:

I have explored creating a TUI browser using Textual to create the TUI and using Go to handle all of the backend logic. I have implemented bm25 ranking for results.
Some limitations include ranking is not perfect, as I do not implement any advanced ranking algorithms beyond bm25. 
Another limitation includes the lack of sources. I have to curate a list of rss feeds manually, therefore I only selected a small number of sites just to test with. 
