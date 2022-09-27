#!/bin/bash
# delete old files
find /home/labuser/script -name "log_*.txt" -type f -mtime +2 -delete 
# by default, cron uses the root directory
echo -e "$(date): starting... \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
# set the last run time
last_run_time=''
if test -f /home/labuser/script/last_run_time; then
    last_run_time=$(sudo cat /home/labuser/script/last_run_time)
    echo "last run time is: $last_run_time \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
    if [ -z "$last_run_time" ]; then
        echo "last run time is empty. set to previous day \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
        last_run_time=$(date  --date="yesterday" +%s)
    fi
else
    # set the last run time to previous day
    echo "last run time is empty. set to previous day \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
    last_run_time=$(date  --date="yesterday" +%s)
fi
# get all the secrets created after the last successful run of the script
# filter the secrets based on the issuer kind and group
kubectl get secret -o json| jq -r "[.items[].metadata]" |jq '[.[] | select(.annotations."cert-manager.io/issuer-kind"=="ClusterIssuer" and .annotations."cert-manager.io/issuer-group"=="cert-manager.io")]'| jq "[.[] |{name: .name, creationTimestamp: .creationTimestamp | fromdate}|  select(.creationTimestamp > $last_run_time)]"| jq -r ".[].name" > /home/labuser/script/secrets.txt
echo "List of new certs: " >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
cat /home/labuser/script/secrets.txt >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
if [ -s /home/labuser/script/secrets.txt ]; then
    #get the KMRA server pid
    pid=`pidof python3.8 apphsm.py`
    if [ $pid > 0  ]; then 
        echo "KMRA is running.. Stop the server \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
        # stop the KMRA server
        sudo kill `pidof python3.8 apphsm.py`
    else
        echo "KMRA is not running \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
    fi
    sleep 5
    #process each secret
    while read secret;  do 
        echo "processing secret $secret \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
        issuer_name=$(kubectl get secret $secret -o jsonpath='{.metadata.annotations.issuer-name}')
        issuer_namespace=$(kubectl get secret $secret -o jsonpath='{.metadata.annotations.issuer-namespace}')
        # generate a unique id for the cert
        crt_id=''
        if [ -z "$issuer_namespace" ]; then
            # this is not a namespaced issuer, set it with the default namespace
            echo "no namespace is defined for the issuer $issuer_name . set the issuer with default namespace \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
            crt_id="tcsissuer.tcs.intel.com/default.$issuer_name"    
        else
        # namespaced issuer, use the namespace.issuername
            echo "set the issue with the namespace $issuer_namespace \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
            crt_id="tcsissuer.tcs.intel.com/$issuer_namespace.$issuer_name"
        fi
        echo "id is: $crt_id \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
        # check a cert with the same id already exists in the apphsm.conf
        length=$(jq .keys /opt/intel/apphsm/apphsm.conf| jq --arg k id --arg v $crt_id  'map(select(.[$k] == $v))' | jq length)
        if [ $length -gt 0 ]; then
            echo -e "$(date): skipping... a cert entry already exists in the apphsm.conf with the id $cert_id \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
            continue
        fi
        # get the key and cert from the secret
        kubectl get secret $secret -o jsonpath='{.data.tls\.key}' |base64 -d > /home/labuser/script/$issuer_name.pem
        kubectl get secret $secret -o jsonpath='{.data.tls\.crt}' |base64 -d > /home/labuser/script/$issuer_name.crt
        #set the key label and key token
        label=$(echo $RANDOM | md5sum | head -c 10; echo;) 
        token_label=$label"-token"
        key_label=$label"-key"
        key=/home/labuser/script/$issuer_name.pem
        cert=/home/labuser/script/$issuer_name.crt
        echo "token label : $token_label \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
        echo "key label : $key_label \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
        echo "key : $key \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
        echo "cert : $cert \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
        #import the key into the KMRA server using the sample key gen utility
        /opt/intel/sample_p11_keygen/sample_key_gen --so-pin 5678 --pin 5678 --token-label $token_label --key-label $key_label --import-key $key
        # update the apphsm.conf
        jq --arg crt_id $crt_id --arg token_name $token_label --arg key_name $key_label --arg certificate_file $cert '.keys[.keys | length] |= . + { "id": $crt_id, "token_name": $token_name, "pin": "5678", "key_name": $key_name, "certificate_file": $certificate_file}' /opt/intel/apphsm/apphsm.conf | ex -sc 'wq!/opt/intel/apphsm/apphsm.conf' /dev/stdin
    done < /home/labuser/script/secrets.txt
    # set  the last run time. save it in to file
    last_run_time=$(date +%s)
    echo "last runtime is $last_run_time \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
    echo $last_run_time > /home/labuser/script/last_run_time
    # start the KMRA server
    sudo make --directory /opt/intel/apphsm/ run &
    sleep 5
else
    echo -e "$(date): skipping... no new secrets \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
    # set  the last run time. save it in to file
    last_run_time=$(date +%s)
    echo "last runtime is $last_run_time \n" >> /home/labuser/script/log_$(date '+%Y-%m-%d').txt
    echo $last_run_time > /home/labuser/script/last_run_time
fi
