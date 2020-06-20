# extractMirthData

[![Build Status](https://travis-ci.org/speedyhoon/extractMirthData.svg?branch=master)](https://travis-ci.org/speedyhoon/extractMirthData)
[![Go Report Card](https://goreportcard.com/badge/github.com/speedyhoon/extractMirthData)](https://goreportcard.com/report/github.com/speedyhoon/extractMirthData)

Parses an XML list of channels exported from Mirth and generates a CSV file. Handy for visualising/maintaining thousands of Mirth channels over hundereds of servers.

1. Open **Mirth Channels** screen
2. Click on **Export All Channels**
3. Run from the command line:
```bat
extractMirthData -xmlDir C:\path_to_new_mirth_export > output.csv
```
