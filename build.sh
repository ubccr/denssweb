#!/bin/bash

#==============================================================================
# DENSSWeb build script
#==============================================================================
#

VERSION=$(git describe --long --tags --dirty --always 2>/dev/null | cut -f2 -d'v')

case "$1" in
        dist)
            mkdir -p dist
            cp -R ./templates dist
            cp -R ./LiteMol/dist/css/* ./dist/templates/static/css/
            cp -R ./LiteMol/dist/js/*.js ./dist/templates/static/js/
            cp -R ./LiteMol/dist/fonts ./dist/templates/static/
            ;;
        *)
            echo $"Usage: $0 {dist|release}"
            exit 1
 
esac
