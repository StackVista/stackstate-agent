#!/bin/bash

FOLDER_PATH="./sut/yards"
FOLDERS=$(find "$FOLDER_PATH" -type d -maxdepth 1 -mindepth 1)

echo ""
echo "Which Yard do you want to use the context for K9S:"
echo ""

counter=1
for folder in $FOLDERS; do
    echo "$counter) $folder"
    ((counter++))
done
echo ""

read -p "Select Yard: " CHOICE

SELECTED_FOLDER=$(echo "$FOLDERS" | sed "${CHOICE}q;d")

k9s --kubeconfig "$SELECTED_FOLDER"/config

