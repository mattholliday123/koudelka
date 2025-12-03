(ALPHA)

koudelka is a CLI tool that fetches the wikipedia summary and sections, if desired, straight to the terminal as well as fetches news from The Guardian. 


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
Future Plans:
Search feature
more options for news
Textual application with vim motions to move through pages and text
