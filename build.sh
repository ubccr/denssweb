#!/bin/bash

#==============================================================================
# DENSSWeb build script
#==============================================================================
#

DENSSWEB_DIR='./.denssweb-release'
VERSION=$(git describe --long --tags --dirty --always 2>/dev/null | cut -f2 -d'v')

# Copy denssweb templates and LiteMol js/css dist/templates
tmpl_dist(){
    mkdir -p dist
    cp -R ./templates dist
    cp -R ./LiteMol/dist/css/* ./dist/templates/static/css/
    cp -R ./LiteMol/dist/js/*.js ./dist/templates/static/js/
    cp -R ./LiteMol/dist/fonts ./dist/templates/static/
}

# Create a denssweb release
make_release(){
    local NAME=denssweb-${VERSION}-${GOOS}-${GOARCH}
    local REL_DIR=${DENSSWEB_DIR}/${NAME}
    go build -ldflags "-X main.Version=$VERSION" .
    rm -Rf ${DENSSWEB_DIR}
    mkdir -p ${REL_DIR}
    cp ./README.rst ${REL_DIR}/ 
    cp ./AUTHORS.rst ${REL_DIR}/ 
    cp ./ChangeLog.rst ${REL_DIR}/ 
    cp ./LICENSE ${REL_DIR}/ 
    cp -R ./dist/templates ${REL_DIR}/ 
    cp -R ./scripts ${REL_DIR}/ 
    cp -R ./ddl ${REL_DIR}/ 

    if [ "$GOOS" == "windows" ]; then
        cp ./denssweb.exe ${REL_DIR}/ 
        cd ${DENSSWEB_DIR}
        zip -r ${NAME}.zip ${NAME}
        mv ${NAME}.zip ../
    else
        cp ./denssweb ${REL_DIR}/ 
        cd ${DENSSWEB_DIR}
        tar cvzf ${NAME}.tar.gz ${NAME}
        mv ${NAME}.tar.gz ../
    fi
    cd ../
    rm -Rf ${DENSSWEB_DIR}
    rm -f ./denssweb
    rm -f ./denssweb.exe
}

case "$1" in
        dist)
            tmpl_dist
            ;;
        release-linux)
            tmpl_dist
            export GOOS=linux
            export GOARCH=amd64
            make_release
            ;;
        release-darwin)
            tmpl_dist
            export GOOS=darwin
            export GOARCH=amd64
            make_release
            ;;
        release-windows)
            tmpl_dist
            export GOOS=windows
            export GOARCH=amd64
            export CGO_ENABLED=1
            export CC=x86_64-w64-mingw32-gcc
            make_release
            ;;
        *)
            echo $"Usage: $0 {dist|release}"
            exit 1
esac
