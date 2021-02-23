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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	GNOMHeaderPattern           = regexp.MustCompile(`^\s+####\s+G N O M[\s\-]+Version\s([0-9]+\.[0-9]+)`)
	GNOMDmaxPattern             = regexp.MustCompile(`^\s+Maximum characteristic size:\s+(\d+\.\d+)`)
	GNOMScatteringHeaderPattern = regexp.MustCompile(`^\s*S\s+J EXP\s+ERROR\s+J REG\s+I REG`)
)

// Validate data of fit file. First column is q, second column is intensity,
// third column is error, fourth column is fit. Returns number of columns or
// error
func validateDAT(data []byte) (int, error) {
	contentType := http.DetectContentType(data)
	if !strings.HasPrefix(contentType, "text/plain") {
		log.WithFields(log.Fields{
			"contentType": contentType,
		}).Error("Invalid input file uploaded")
		return 0, fmt.Errorf("Invalid input data. Please provide an ascii text file")
	}

	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)
	lineno := 0
	cols := 0
	isEmpty := true
	for scanner.Scan() {
		lineno++
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			// skip blank lines
			continue
		}
		if strings.HasPrefix(line, "#") {
			// skip comments
			continue
		}
		parts := strings.Fields(line)
		for _, n := range parts {
			_, err := strconv.ParseFloat(n, 64)
			if err != nil {
				return 0, fmt.Errorf("Invalid floating point numbers found on line %d", lineno)
			}
			isEmpty = false
		}

		if len(parts) > 0 {
			cols = len(parts)
		}
	}

	if isEmpty {
		return 0, errors.New("Input data file was empty")
	}

	return cols, nil
}

// Check if input data has GNOM header and return version
func parseGNOMHeader(data []byte) (float64, error) {
	contentType := http.DetectContentType(data)
	if !strings.HasPrefix(contentType, "text/plain") {
		log.WithFields(log.Fields{
			"contentType": contentType,
		}).Error("Invalid GNOME file uploaded")
		return 0, fmt.Errorf("Invalid GNOME input file not an ascii text file")
	}

	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)
	lineno := 0
	for scanner.Scan() {
		lineno++
		line := scanner.Text()

		matches := GNOMHeaderPattern.FindStringSubmatch(line)
		if len(matches) == 2 {
			version, err := strconv.ParseFloat(matches[1], 64)
			if err == nil {
				return version, nil
			}
		}

		// Only check first 20 lines
		if lineno > 20 {
			break
		}
	}

	return 0, errors.New("Input data is not GNOM")
}

// Convert GNOM formatted data into simple 3-column DAT file. First column is
// q, second column is intensity, third column is error. This assumes
// parseGNOMHeader has already been called. Returns the converted data and Dmax
func convertGNOM(data []byte, version float64) ([]byte, float64, error) {
	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)
	lineno := 0
	dmax := float64(0)
	headerFound := false
	// Scan for scattering data header and Dmax (if v5)
	for scanner.Scan() {
		lineno++
		line := scanner.Text()
		if matches := GNOMDmaxPattern.FindStringSubmatch(line); len(matches) == 2 {
			dmax, _ = strconv.ParseFloat(matches[1], 64)
		} else if GNOMScatteringHeaderPattern.MatchString(line) {
			headerFound = true
			break
		}
	}

	if !headerFound {
		return nil, 0, errors.New("Failed to find Scattering data header line in GNOM file")
	}

	records := [][]string{}
	// At the beginning the extrapolated data is only 2 columns. So we use the
	// first error (3rd column) value we see and use it to fill in the missing data
	firstValueFor3rdColumn := ""
	for scanner.Scan() {
		lineno++
		line := scanner.Text()
		line = strings.TrimSpace(line)
		// skip header and blank lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for end of scattering data
		if strings.Contains(line, "function of particle") || // gnom jobtype 0
			strings.Contains(line, "particle thickness") || // gnom jobtype 3
			strings.Contains(line, "function of cross-section") || // gnom jobtype 4
			strings.Contains(line, "function of long cylinders") || // gnom jobtype 5
			strings.Contains(line, "function of spherical shells") { // gnom jobtype 6
			break
		}

		parts := strings.Fields(line)

		// Ensure floating point numbers
		for _, n := range parts {
			_, err := strconv.ParseFloat(n, 64)
			if err != nil {
				return nil, 0, fmt.Errorf("Invalid floating point numbers found on line %d", lineno)
			}
		}

		if len(parts) == 2 {
			records = append(records, []string{parts[0], parts[1], ""})
		} else if len(parts) == 5 {
			records = append(records, []string{parts[0], parts[4], parts[2]})
			if firstValueFor3rdColumn == "" {
				firstValueFor3rdColumn = parts[2]
			}
		} else {
			return nil, 0, fmt.Errorf("Input data format must be 2 or 5 columns: error on line %d", lineno)
		}
	}

	if firstValueFor3rdColumn == "" {
		return nil, 0, fmt.Errorf("Input data only consists of 2 columns")
	}

	if dmax == 0 {
		// Parse Dmax from the Probability data (v4)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			// skip header and blank lines
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Only interested in last column of 3 floats
			parts := strings.Fields(line)
			if len(parts) != 3 {
				continue
			}

			for _, n := range parts {
				_, err := strconv.ParseFloat(n, 64)
				if err != nil {
					continue
				}
			}
			dmax, _ = strconv.ParseFloat(parts[0], 64)
		}
	}

	var buf bytes.Buffer
	for _, rec := range records {
		if rec[2] == "" {
			rec[2] = firstValueFor3rdColumn
		}
		buf.Write([]byte(rec[0]))
		buf.Write([]byte(" "))
		buf.Write([]byte(rec[1]))
		buf.Write([]byte(" "))
		buf.Write([]byte(rec[2]))
		buf.Write([]byte("\n"))
	}

	return buf.Bytes(), dmax, nil
}
