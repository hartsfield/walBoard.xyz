# use in n/vim to restart on save:
# :autocmd BufWritePost * silent! !./autoload.sh app_command
#!/bin/bash
pkill $1 || true
go build -o $1
./$1 >> log.txt 2>&1 &
