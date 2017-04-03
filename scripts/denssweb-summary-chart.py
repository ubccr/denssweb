#!/usr/bin/env python

# Copyright 2017 DENSSWeb Authors. All rights reserved.
# 
# This file is part of DENSSWeb.
# 
# DENSSWeb is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
# 
# DENSSWeb is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
# 
# You should have received a copy of the GNU General Public License
# along with DENSSWeb.  If not, see <http://www.gnu.org/licenses/>.

import argparse
import logging
import sys
import os
import glob
import matplotlib as mpl
import numpy as np
mpl.use('Agg')
import matplotlib.pyplot as plt

# Create summary plot of DENSS run
def summary_plot(work_dir, out_file):
    files = glob.glob(os.path.join(work_dir, 'output_*stats_by_step.dat'))
    shape = np.loadtxt(files[0]).shape
    n = len(files)
    chi2 = np.zeros((n))
    rg = np.zeros((n))
    supportV = np.zeros((n))

    data = np.zeros((n,shape[0],shape[1]))
    for i in range(n):
        data[i] = np.loadtxt(files[i])

    logging.info("Plotting summary chart")
    with plt.style.context('ggplot'):
        f, ax = plt.subplots(3, sharex=True)
        ax[0].set_title('Statistics by Step')
        x = range(shape[0])

        for i in xrange(n):
            ax[0].plot(x, data[i,:,0])
            ax[1].plot(x, data[i,:,1])
            ax[2].plot(x, data[i,:,2])

            chi2[i] = data[i,data[i,:,0]!=0,0][-1]
            rg[i] = data[i,data[i,:,1]!=0,1][-1]
            supportV[i] = data[i,data[i,:,2]!=0,2][-1]

        chi2ave, chi2sd = np.mean(chi2), np.std(chi2)
        rgave, rgsd = np.mean(rg), np.std(rg)
        supportVave, supportVsd = np.mean(supportV), np.std(supportV)

        ax[0].semilogy()
        ax[0].set_ylabel(r'$\chi^2$')
        ax[0].text(0.5, 0.8, r'Average $={:.3f}$ $\sigma={:.3f}$'.format(chi2ave, chi2sd), transform=ax[0].transAxes)

        ax[1].set_ylabel(r'$R_g$')
        ax[1].text(0.5, 0.8, r'Average $={:.3f}$ $\sigma={:.3f}$'.format(rgave, rgsd), transform=ax[1].transAxes)

        ax[2].semilogy()
        ax[2].set_ylabel(r'Support Volume')
        ax[2].text(0.5, 0.8, r'Average $={:.3f}$ $\sigma={:.3f}$'.format(supportVave, supportVsd), transform=ax[2].transAxes)

        plt.savefig(out_file, format='png', bbox_inches="tight", pad_inches=0.5)

def main():
    logging.basicConfig(
        format='%(asctime)s [%(levelname)s] %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S',
        level=logging.CRITICAL
    )

    parser = argparse.ArgumentParser(description='Create summary plot of DENSS run for use with DENSSWeb')
    parser.add_argument("-v", "--verbose", help="output debugging information", action="store_true")
    parser.add_argument("-i", "--input", help="Path to work directory (/workdir/denss-1)")
    parser.add_argument("-o", "--output", help="Path to output file")

    args = parser.parse_args()
    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)

    if not args.input or not args.output:
        logging.critical("Please specify both a work directory and output file")
        sys.exit(1)

    summary_plot(args.input, args.output)

if __name__ == "__main__":
    main()
