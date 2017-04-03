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
import matplotlib as mpl
mpl.use('Agg')
import matplotlib.pyplot as plt

# Create Fourier Shell Correlation (FSC) curve
def fsc_plot(in_file, out_file):
    x = []
    y = []

    logging.info("Parsing input file")
    with open(in_file, 'r') as fh:
        for line in fh:
            xi, yi = line.strip().split()
            x.append(float(xi))
            y.append(float(yi))


    resolution = 0
    cutoff = 0.5

    logging.info("Estimate resolution")
    for i in xrange(len(x)-1, 0, -1):
        if y[i] > 0.5:
            resolution = x[i]
            break

    if resolution > 0:
        resolution = 1/resolution

    logging.info("Plotting fsc curve")
    with plt.style.context('ggplot'):
        fig, ax = plt.subplots()

        ax.plot(x, y)
        ax.text(0.8, 0.9, r'$r={:.3f} \AA$'.format(resolution), transform=ax.transAxes)

        plt.axhline(0.5, color='#444444', linewidth=1, linestyle='dashed')
        plt.axhline(0.143, color='#444444', linewidth=1, linestyle='dashed')
        plt.xlabel(r'Resolution (1/$\AA$)')
        plt.ylabel('FSC')
        plt.savefig(out_file, format='png', bbox_inches="tight", pad_inches=0.5)

def main():
    logging.basicConfig(
        format='%(asctime)s [%(levelname)s] %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S',
        level=logging.CRITICAL
    )

    parser = argparse.ArgumentParser(description='Create Fourier Shell Correlation for use with DENSSWeb')
    parser.add_argument("-v", "--verbose", help="output debugging information", action="store_true")
    parser.add_argument("-i", "--input", help="Path to input file (spt_01/fsc_0.txt)")
    parser.add_argument("-o", "--output", help="Path to output file")

    args = parser.parse_args()
    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)

    if not args.input or not args.output:
        logging.critical("Please specify both an input and output file")
        sys.exit(1)

    fsc_plot(args.input, args.output)

if __name__ == "__main__":
    main()
