# Installations

install python3
install flask
install libsqlite3

# Refactoring

Control.sh
Set the first line to #!/bin/bash
Update the path for DATABASE from '/tmp/minitwit.db' to ‘./minitwit.db’
Set “$” around all variables, i.e from $1 to "$1" 
Update "$(which python)" to "(which python)" as the variable has not been set

minitwit.py
Run the command: 2to3 minitwit.py -w -n -o minitwit2.py
Delete the python2 version and rename the python3 version to minitwit.py
Remove initial comments and utf-8 coding 
Update werkzeug to werkzeug.security
Update the path for DATABASE to ‘./minitwit.db’
Set the mode for app.open_resource(‘schema.sql’, mode=‘r’)

minitwit_tests.py 
Put b in front of all asserted terms, i.e. assert b’You were logged in’ in rv.data

# Compiling the code

gcc flag_tool.c -l:libsqlite3.a -lm -o flag_tool  
python3 minitwit.py 
python3 minitwit_tests.py
