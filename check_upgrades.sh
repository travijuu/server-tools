#!/bin/bash

#
# settings
#
ENDPOINT=""
CHANNEL=""
username=$(hostname)

#
# check available OS upgrades
#
apt-get update
mapfile -t packages < <(aptitude search '~U')

total=${#packages[*]}
for (( i=0; i<$total; i++ ))
do
    if [[ "${packages[$i]}" == *"linux-headers-generic"* ]] ||
       [[ "${packages[$i]}" == *"linux-generic"* ]] ||
       [[ "${packages[$i]}" == *"linux-image-generic"* ]]
    then
       unset packages[$i]
    fi
done

if [[ ${#packages[*]} -eq 0  ]]; then
    exit;
fi

#
# Prepare POST data
#
message=$( IFS=$'\n'; echo "${packages[*]}" )
message="The list of available Linux updates \`\`\`$message\`\`\`"
json="{\"channel\": \"#$CHANNEL\", \"username\":\"$username\", \"mrkdwn\":true, \"text\": \"${message}\"}"

#
# POST to Slack
#
curl -s -d "payload=$json" "$ENDPOINT"
