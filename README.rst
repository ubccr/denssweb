===============================================================================
DENSSWeb - Web frontend to DENSS
===============================================================================

DENSSWeb is the web frontend for the `DENSS algorithm <https://github.com/tdgrant1/denss>`_.

------------------------------------------------------------------------
Installation
------------------------------------------------------------------------

TODO

------------------------------------------------------------------------
Building from source
------------------------------------------------------------------------

TODO

------------------------------------------------------------------------
Required Software
------------------------------------------------------------------------

DENSSWeb requires the following software

* DENSS
* EMAN2
* Situs

The following assumes you're running Ubuntu 16.04. Install required
packages::

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
    $ pyton setup.py install

Installing EMAN2
~~~~~~~~~~~~~~~~~

Clone eman2 source code::

    $ git clone https://github.com/cryoem/eman2
    $ cd eman2

Run cmake::

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

Installing Situs
~~~~~~~~~~~~~~~~~~~~~~~~~~~

You can fetch Situs source from `here <http://situs.biomachina.org/>`_. Only
really need the map2map command. 

Compile and install::

    $ tar xvzf Situs_2.8.tar
    $ cd src
    $ make
    $ make install

------------------------------------------------------------------------
License
------------------------------------------------------------------------

DENSSWeb is released under the GPLv3 license. See the LICENSE file.

