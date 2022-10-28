#!/usr/bin/bash

go build ../../
read -p "User: " user
read -s -p "Password: " password
echo ""
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/defPpMinimal.spec

./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2007.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2008.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2009.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2010.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2011.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2012.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2013.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2014.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2015.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2016.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2017.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2018.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2019.spec
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPp/defPp2020.spec

