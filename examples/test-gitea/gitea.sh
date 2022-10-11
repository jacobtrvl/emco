#!/bin/bash

function get_variables {
    user_name=emco
    ip=$HOST_IP
    port=$(kubectl get svc -n emco | grep gitea-http | awk -F[\:\/] '{print $2}')
    password=emco@pass
    token_name="gitea_token"
    url="http://"$ip":"$port
}

function get_token {
token=$(curl -H "Content-Type: application/json" -d "{\"name\":\"$token_name\"}" -u $user_name:$password $url/api/v1/users/$user_name/tokens | jq -r '.sha1')
}

function create_repo {
read -p 'UserName: ' user_name
read -sp 'Password: ' password; echo
read -p 'RepoName: ' repo_name
token_name="gitea_token"
get_token
end_point=$url"/api/v1/user/repos"
output=$(curl -k -H "content-type: application/json" -H "Authorization: token ${token}" --data "{\"name\":\"$repo_name\",\"auto_init\":true,\"description\":\"Test Repo\",\"gitingores\":\"\",\"readme\":\"\"}" "${end_point}")

created=$( printf '%s' "$output" | jq -r '.created_at')

if [ "$created" == "null" ]
then
    echo "Error in creating repo, check error message"
    echo $output
else
    echo "New Repo Created! Access using URL: " $url
fi

delete_token
}

function delete_token {
    end_point=$url"/api/v1/users/"$user_name"/tokens/"$token_name
    cred=$user_name":"$password
    cred_base64="$(echo -n $cred | base64)"
    curl -X 'DELETE' "$end_point" -H 'accept: application/json' -H "authorization: Basic $cred_base64"
}

function create_user {
   end_point=$url"/api/v1/admin/users"
   read -p 'UserName: ' uservar
   read -p 'Email: ' emailvar
   read -sp 'Password: ' passvar; echo
   cred=$user_name":"$password
   cred_base64="$(echo -n $cred | base64)"
   output=$(curl -X 'POST'   "$end_point"   -H 'accept: application/json'   -H "authorization: Basic $cred_base64"   -H 'Content-Type: application/json'   -d "{
  \"email\": \"$emailvar\",
  \"full_name\": \"$uservar\",
  \"login_name\": \"$uservar\",
  \"must_change_password\": false,
  \"password\": \"$passvar\",
  \"restricted\": true,
  \"send_notify\": true,
  \"source_id\": 0,
  \"username\": \"$uservar\",
  \"visibility\": \"\"
}")

created=$( printf '%s' "$output" | jq -r '.created')

if [ "$created" == "null" ]
then
    echo "Error in creating user, check error message"
    echo $output
else
    echo "New user Created! Access using URL: " $url
fi
}


function usage {

    echo "Usage: $0 create_user|create_repo"
    echo "Example: $0 create_user"

}

get_variables


if [ "$#" -lt 1 ] ; then
    usage
    exit
fi

case "$1" in
    "create_user" ) create_user ;;
    "create_repo" ) create_repo ;;
    *) usage ;;
esac
