===============================================================================
Integrated LiteMol plugin for DENSS
===============================================================================

DENSSWeb provides an integrated `LiteMol <https://github.com/dsehnal/litemol>`_
plugin for displaying electron density maps (DENSS Viewer). The DENSS viewer
plugin is written in TypeScript and uses the LiteMol Plugin API.

------------------------------------------------------------------------
Compiling the plugin
------------------------------------------------------------------------

You'll need to first compile LiteMol (included as a git submodule). This
assumes you have nodejs installed (tested with v6.10.0)::

    $ cd ../LiteMol
    $ npm install -g gulp
    $ npm install
    $ gulp

To compile the DENSS viewer LiteMol plugin run (from this directory)::

    $ tsc

This will compile the DENSS viewer plugin and place the js code in
../templates/static/js/LiteMol-denss.js. See tsconfig.json for details
