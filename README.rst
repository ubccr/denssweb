===============================================================================
DENSSWeb - Web frontend to DENSS
===============================================================================

DENSSWeb is a web frontend for the `DENSS algorithm <https://github.com/tdgrant1/denss>`_.
DENsity from Solution Scattering (DENSS) is an algorithm used for calculating
ab initio electron density maps directly from solution scattering data.
DENSSWeb provides a web based frontend interface allowing users to run complex
DENSS pipelines and view results in a browser. DENSSWeb performs parallel DENSS
runs and averages the results using EMAN2. The resulting density map is
displayed using a custom LiteMol plugin along with a plot of the fourier shell
correlation (FSC) curve.

A demo fo DENSSWeb can be found `here <https://denss.ccr.buffalo.edu>`_.

DENSSWeb can be run locally on a single machine or on multiple machines in a
clustered environment. DENSSWeb consists of a server and client worker. The
server runs an embedded web server and the client worker runs the DENSS
pipeline.

------------------------------------------------------------------------
Requirements
------------------------------------------------------------------------

A web browser with WebGL support. To check if your browser supports WebGL `see
here <https://get.webgl.org/>`_.

------------------------------------------------------------------------
Install
------------------------------------------------------------------------

Install all required software (see Required Software section). Download the
current release of DENSSWeb for your platform `here <https://github.com/ubccr/denssweb/releases>`_.

Unpack the DENSSWeb release::

    $ tar xzvf denssweb-VERSION-OS-amd64.tar.gz
    $ cd denssweb-VERSION-OS-amd64

Create the config file and edit the paths to required software::

    $ cp denssweb.yaml.sample denssweb.yaml
    (edit to taste)

Start the DENSSWeb client/server::

    $ ./denssweb -d run

Point your browser at http://localhost:8080 and submit a Job

The raw output files for each job are stored in ``work_dir/denss-JOBID``.
``work_dir`` by default is set to a directory named ``denssweb-work`` in your
current working directory but you can override this in the denssweb.yaml file.
The complete log file for a job is in a file named ``denss-JOBID.log``.

If you're running DENSSWeb on a server you must edit the ``bind`` and
``base_url`` settings accordingly.

------------------------------------------------------------------------
Building from source
------------------------------------------------------------------------

Compile the DENSS Viewer LiteMol plugin. See `here <denss-viewer/README.rst>`_

DENSSWeb is written in `Go <https://golang.org/>`_ and uses `glide <http://glide.sh/>`_ 
for package management. Be sure you have a working Go environment and have
glide installed.  To compile from source run::

    $ go get -u github.com/ubccr/denssweb
    $ cd $GOPATH/src/github.com/ubccr/dessweb
    $ ./build.sh tmpl
    $ cp denssweb.yaml.sample denssweb.yaml
    (edit to taste)
    $ denssweb -d run

------------------------------------------------------------------------
Required Software
------------------------------------------------------------------------

DENSSWeb requires the following software
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

* `DENSS <https://github.com/tdgrant1/denss>`_
* `EMAN2 <https://github.com/cryoem/eman2>`_

The following guide assumes you're running Linux Ubuntu 16.04. You will need to
adjust package names for your distro.

Install required packages::

    $ apt-get install libhdf5-10 libhdf5-dev libpng12-0 libpng12-dev libtiff5 libtiff5-dev \
         python-qt4 python-qt4-gl python-opengl python-matplotlib libfftw3-3 libfftw3-dev \
         libgsl0-dev db-util libdb-dev python-bsddb3 libboost-all-dev python-dev cmake \
         cmake-curses-gui ipython libgl1-mesa-dev libglu1-mesa-dev libftgl2 libftgl-dev
         python-scipy build-essential git

Installing DENSS
~~~~~~~~~~~~~~~~~

Clone and install DENSS::

    $ git clone https://github.com/tdgrant1/denss
    $ cd denss
    $ python setup.py install

Installing EMAN2
~~~~~~~~~~~~~~~~~

Clone eman2 source code and run cmake::

    $ git clone https://github.com/cryoem/eman2
    $ cd eman2
    $ mkdir build
    $ cd build
    $ ccmake
    (type c to configure)
    (type e to exit)

Fix the following variables::

    PYTHON_LIBRARY=/usr/lib/python2.7/config-x86_64-linux-gnu/libpython2.7.so
    ZLIB_LIBRARY=/usr/lib/x86_64-linux-gnu/libz.so
    HDF5_INCLUDE_PATH=/usr/include/hdf5/serial
    HDF5_LIBRARY=/usr/lib/x86_64-linux-gnu/hdf5/serial/libhdf5.so
    TIFF_INCLUDE_PATH=/usr/include/x86_64-linux-gnu

Compile and install::

    $ make
    $ make install

Setup env variables in ~/.bashrc::

    export EMAN2DIR=$HOME/EMAN2
    export PATH=$PATH:$EMAN2DIR/bin
    export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$EMAN2DIR/lib
    export PYTHONPATH=$PYTHONPATH:$EMAN2DIR/lib

------------------------------------------------------------------------
License
------------------------------------------------------------------------

DENSSWeb is released under the GPLv3 license. See the LICENSE file.
