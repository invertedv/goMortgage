#!/usr/bin/fish

go build ../../
read -p "User: " user
read -s -p "Password: " password
echo ""

cd ../sequentialDefPp
./defPpSeq.sh

./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/defPpFull.gom

./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2007.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2008.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2009.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2010.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2011.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2012.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2013.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2014.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2015.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2016.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2017.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2018.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2019.gom
./goMortgage -user $user -pw $password -specs /home/will/GolandProjects/goMortgage/scripts/sequentialDefPpFull/defPpFull2020.gom

