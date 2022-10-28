#!/usr/bin/bash

go build ../../
read -p "User: " user
read -s -p "Password: " password
echo ""
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/mod.gom

./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2007.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2008.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2009.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2010.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2011.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2012.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2013.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2014.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2015.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2016.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2017.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2018.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2019.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialMod/mod2020.gom

