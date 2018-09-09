#!/bin/bash
echo -n "Run deployment? (y/n)? "
old_stty_cfg=$(stty -g)
stty raw -echo
answer=$( while ! head -c 1 | grep -i '[ny]' ;do true ;done )
stty $old_stty_cfg
if echo "$answer" | grep -iq "^y" ;then
    echo
    echo
    echo "Running Gulp..."
    gulp buildfordeploy
    echo
    echo "Deploying indexes..."
    gcloud app deploy --quiet main/index.yaml
    echo
    echo "Deploying application..."
    gcloud app deploy --quiet main
    echo
    echo "Deployment script complete."
else
    echo ""
fi