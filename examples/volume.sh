#!/bin/bash
set -e

function volume () {
  amixer get Master | grep '%' | head -n 1 | cut -d '[' -f 2 | cut -d '%' -f 1
}

function is_mute () {
  amixer get Master | grep '%' | grep -oE '[^ ]+$' | grep off > /dev/null
}

function volume_icon () {
  volume=$(volume)
  if is_mute; then
    echo ""
  elif [[ "$volume" = "0" ]]; then
    echo ""
  elif [[ volume -lt 50 ]]; then
    echo ""
  else
    echo ""
  fi
}

function volume_percentage () {
  volume=$(volume)
  if is_mute; then
    echo "M"
  else
    echo "$volume%"
  fi
}

if [[ "$1" == "icon" ]]; then
  volume_icon
elif [[ "$1" == "percentage" ]]; then
  volume_percentage
fi
