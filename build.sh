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
    for os in linux
    do
        for arch in amd64
        do
            NAME=denssweb-${VERSION}-${os}-${arch}
            REL_DIR=${DENSSWEB_DIR}/${NAME}
            GOOS=$os GOARCH=$arch go build -ldflags "-X main.Version=$VERSION" .
            rm -Rf ${DENSSWEB_DIR}
            mkdir -p ${REL_DIR}
            cp ./denssweb ${REL_DIR}/ 
            cp ./README.rst ${REL_DIR}/ 
            cp ./AUTHORS.rst ${REL_DIR}/ 
            cp ./ChangeLog.rst ${REL_DIR}/ 
            cp ./LICENSE ${REL_DIR}/ 
            cp -R ./dist/templates ${REL_DIR}/ 
            cp -R ./scripts ${REL_DIR}/ 
            cp -R ./ddl ${REL_DIR}/ 

            cd ${DENSSWEB_DIR} && zip -r ${NAME}.zip ${NAME}
            mv  ${NAME}.zip ../
            cd ../
            rm -Rf ${DENSSWEB_DIR}
            rm ./denssweb
        done
    done
}

case "$1" in
        dist)
            tmpl_dist
            ;;
        release)
            tmpl_dist
            make_release
            ;;
        *)
            echo $"Usage: $0 {dist|release}"
            exit 1
esac
