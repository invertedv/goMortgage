#!/usr/bin/fish
go build .

read -p "User: " user
read -s -p "Password: " password
echo ""

./mortgage -user $user -pw $password /home/will/GolandProjects/mortgage/scripts/dq.gom
