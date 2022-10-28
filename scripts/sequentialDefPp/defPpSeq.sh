#!/usr/bin/bash

go build ../../
read -p "User: " user
read -s -p "Password: " password
echo ""
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/defPpMinimal.gom

./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2007.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2008.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2009.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2010.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2011.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2012.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2013.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2014.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2015.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2016.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2017.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2018.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2019.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2020.gom

