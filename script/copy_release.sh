#!/bin/bash

OLD_RELEASE=scumbago
NEW_RELEASE=scumbago.new

read -p "Copy new release? (y/N) " -n 1 -r
if [[ $REPLY =~ ^[Yy]$ ]]; then
  echo

  if [ ! -f $NEW_RELEASE ]; then
    echo "$NEW_RELEASE file not found"
    exit 1
  fi

  VERSION=`./$OLD_RELEASE -version | cut -d ' ' -f 2`
  cp -v $OLD_RELEASE $OLD_RELEASE-$VERSION
  mv -fv $NEW_RELEASE $OLD_RELEASE
fi
