#!/usr/bin/bash

go build ../../
read -p "User: " user
read -s -p "Password: " password
echo ""
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/mod.spec

./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2007.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2008.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2009.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2010.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2011.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2012.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2013.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2014.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2015.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2016.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2017.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2018.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2019.spec
./mortgage -user $user -pw $password -specs /home/will/GolandProjects/mortgage/scripts/sequentialMod/mod2020.spec

