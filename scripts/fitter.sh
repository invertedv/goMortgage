#!/usr/bin/bash
go build ../

read -p "User: " user
read -s -p "Password: " password
echo ""

#./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/prepayScore.gom

#./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/mod.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/allInEven.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/allInEvenStrat.gom
#./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/allInStrat.gom

