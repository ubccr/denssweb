// Copyright 2017 DENSSWeb Authors. All rights reserved.
//
// This file is part of DENSSWeb.
//
// DENSSWeb is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// DENSSWeb is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with DENSSWeb.  If not, see <http://www.gnu.org/licenses/>.

package server

import (
	"bytes"
	"testing"
)

func TestDAT(t *testing.T) {
	good_data := `
0 68517072 14388.6
0.001 68512672 22609.2 
0.002 68499000 33907
0.003 68475592 49302.4

0.004 68442896 61598.6
0.005 68400936 73873
0.006 68349816 71767.3
0.007 68289288 81947.1
0.008 68219544 110516
0.009 68140592 132874
0.01 68052672 134744 
0.011 67955744 134552
    `

	err := validateDAT([]byte(good_data))
	if err != nil {
		t.Fatal(err)
	}

	bad_data := `
0.003 68475592 49302.4
THIS IS NOT ALLOWED
0.004 68442896 61598.6
    `

	err = validateDAT([]byte(bad_data))
	if err == nil {
		t.Errorf("Invalid DAT provided")
	}
}

func TestGNOMHeader(t *testing.T) {
	headers := map[float64]string{
		float64(5.0): `
           ####      G N O M              Version 5.0 (r8972)      ####
                                               Wed Mar  8 16:11:42 2017
    
           ####      Configuration                                 ####

    System Type:                   arbitrary monodisperse (job = 0)`,
		float64(4.6): `

           ####    G N O M   ---   Version 4.6                       ####

                                                    04-May-2016   15:51:44
           ===    Run No   1   ===
 Run title:  ### DATA:


   *******    Input file(s) : t_dat.dat
           Condition P(rmin) = 0 is used.
           Condition P(rmax) = 0 is used.`}

	for v, header := range headers {
		version, err := parseGNOMHeader([]byte(header))
		if err != nil {
			t.Fatal(err)
		}

		if version != v {
			t.Errorf("Incorrect GNOM version parsed: got %.2f should be %.2f", version, v)
		}
	}
}

func TestConvertGNOM(t *testing.T) {
	input := map[float64][]string{
		float64(50.0): []string{`

           ####      G N O M              Version 5.0 (r8972)      ####
                                               Wed Mar  8 16:11:42 2017

           ####      Configuration                                 ####

    System Type:                   arbitrary monodisperse (job = 0)
    Minimum characteristic size:         0.0000
    Maximum characteristic size:        50.0000
    rad56:                               0.0000
    Force 0.0 at r = rmin:                  yes
    Force 0.0 at r = rmax:                  yes
    Initial alpha:                       0.7930
    Initial random seed:    5349875249374940886
    Points in real space:                   256

    Input  1:                          6lyz.dat
    First data point used:                    1
    Last data point used:                  6301
    Scaling coefficient:             0.1000E+01
    Experimental setup:            point collimation


           ####      Results                                       ####

    Parameter    DISCRP  OSCILL  STABIL  SYSDEV  POSITV  VALCEN  SMOOTH
    Weight        1.000   3.000   3.000   3.000   1.000   1.000   2.000
    Sigma         0.300   0.600   0.120   0.120   0.120   0.120   0.600
    Ideal         0.700   1.100   0.000   1.000   1.000   0.950   0.000
    Current       0.003   1.462   0.000   0.301   1.000   0.919   0.004
               --------------------------------------------------------
    Estimate      0.005   0.695   1.000   0.000   1.000   0.935   1.000

    Angular range:                       0.0000 to       6.3000
    Reciprocal space Rg:             0.1408E+02
    Reciprocal space I(0):           0.6852E+08

    Real space range:                    0.0000 to      50.0000
    Real space Rg:                   0.1429E+02 +-   0.3400E-01
    Real space I(0):                 0.6852E+08 +-   0.1423E+05

    Highest ALPHA (theor):           0.2529E+10
    Current ALPHA:                   0.7930E+00

    Total Estimate:                      0.6447 (a REASONABLE solution)



           ####      Experimental Data and Fit                     ####

      S          J EXP       ERROR       J REG       I REG

   0.000000E+00   0.685171E+08   0.143886E+05   0.685175E+08   0.685175E+08
   0.100000E-02   0.685127E+08   0.226092E+05   0.685127E+08   0.685127E+08
   0.200000E-02   0.684990E+08   0.339070E+05   0.684987E+08   0.684987E+08
   0.630000E+01   0.841165E+01   0.874530E+01   0.840905E+01   0.840905E+01

           ####      Real Space Data                               ####

           Distance distribution  function of particle


      R          P(R)      ERROR

  0.0000E+00  0.0000E+00  0.0000E+00
    `,
			`0.000000E+00 0.685175E+08 0.143886E+05
0.100000E-02 0.685127E+08 0.226092E+05
0.200000E-02 0.684987E+08 0.339070E+05
0.630000E+01 0.840905E+01 0.874530E+01
`},
		float64(101.0): []string{`
           ####    G N O M   ---   Version 4.6                       ####

                                                    04-May-2016   15:51:44
           ===    Run No   1   ===
 Run title:  ### DATA:


   *******    Input file(s) : test.dat
           Condition P(rmin) = 0 is used.
           Condition P(rmax) = 0 is used.

          Highest ALPHA is found to be   0.2000E+03

  The measure of inconsistency AN1 equals to    0.8565E+00
     Alpha    Discrp  Oscill  Stabil  Sysdev  Positv  Valcen    Total
  0.2000E+06 61.5013  1.1246  0.9363  0.0597  1.0000  0.8712  0.38707

             ####            Final results            ####

 Parameter    DISCRP    OSCILL    STABIL    SYSDEV    POSITV    VALCEN
 Weight        1.000     3.000     3.000     3.000     1.000     1.000
 Sigma         0.300     0.600     0.120     0.120     0.120     0.120
 Ideal         0.700     1.100     0.000     1.000     1.000     0.950
 Current       0.464     1.224     0.008     0.950     1.000     0.933
                - - - - - - - - - - - - - - - - - - - - - - - - -
 Estimate      0.539     0.958     0.995     1.000     1.000     0.979

  Angular   range    :     from    0.0097   to    0.2395
  Real space range   :     from      0.00   to    101.00

  Highest ALPHA (theor) :   0.200E+06                 JOB = 0
  Current ALPHA         :   0.521E+01   Rg :  0.334E+02   I(0) :   0.221E+01

           Total  estimate : 0.948  which is  AN EXCELLENT  solution

      S          J EXP       ERROR       J REG       I REG

  0.0000E+00                                      0.2208E+01
  0.5720E-03                                      0.2208E+01
  0.1144E-02                                      0.2207E+01
  0.9725E-02  0.2200E+01  0.2851E-01  0.2132E+01  0.2132E+01
  0.1030E-01  0.2132E+01  0.2473E-01  0.2123E+01  0.2123E+01
  0.2395E+00  0.5065E-02  0.3458E-02  0.5288E-02  0.5288E-02

           Distance distribution  function of particle


       R          P(R)      ERROR

  0.0000E+00  0.0000E+00  0.0000E+00
  0.1010E+01  0.3363E-04  0.2103E-05
  0.9898E+02 -0.9114E-06  0.8662E-05
  0.9999E+02 -0.1002E-05  0.4821E-05
  0.1010E+03  0.0000E+00  0.0000E+00
          Reciprocal space: Rg =   33.39     , I(0) =   0.2208E+01
     Real space: Rg =   33.36 +- 0.060  I(0) =   0.2208E+01 +-  0.3610E-02
         `,
			`0.0000E+00 0.2208E+01 0.2851E-01
0.5720E-03 0.2208E+01 0.2851E-01
0.1144E-02 0.2207E+01 0.2851E-01
0.9725E-02 0.2132E+01 0.2851E-01
0.1030E-01 0.2123E+01 0.2473E-01
0.2395E+00 0.5288E-02 0.3458E-02
`}}

	for dm, d := range input {
		v, _ := parseGNOMHeader([]byte(d[0]))
		data, dmax, err := convertGNOM([]byte(d[0]), v)
		if err != nil {
			t.Errorf(err.Error())
		}

		if bytes.Compare(data, []byte(d[1])) != 0 {
			t.Errorf("Incorrect GNOM data parsed: got \n\n%s\n\n should be \n\n%s\n\n", string(data), d[1])
		}

		if dmax != dm {
			t.Errorf("Incorrect Dmax: got %.3f should be %.3f", dmax, dm)
		}
	}
}
