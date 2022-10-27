#!/usr/bin/bash

go build ../../
read -p "User: " user
read -s -p "Password: " password
echo ""
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/dq.spec

./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2007.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2008.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2009.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2010.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2011.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2012.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2013.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2014.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2015.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2016.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2017.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2018.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2019.spec
./goMortage -user $user -pw $password -specs /home/will/GolandProjects/goMortgagescripts/sequentialDq/dq2020.spec

