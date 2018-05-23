Read a single key from the terminal. It's especially helpful with shell scripts.
For example run this then press a key combination like Ctrl-Alt-j:

    $ echo "You pressed: $(getkey)"
    You pressed: Ctrl-Alt-j
  
## Installation

    go get -u github.com/mndrix/getkey/...
