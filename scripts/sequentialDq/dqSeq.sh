#!/usr/bin/bash

go build ../../
read -p "User: " user
read -s -p "Password: " password
echo ""
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/dq.gom

./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2007.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2008.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2009.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2010.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2011.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2012.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2013.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2014.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2015.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2016.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2017.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2018.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2019.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDq/dq2020.gom

