package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Channel represents a Mirth channel
type Channel struct {
	XMLName     xml.Name  `xml:"channel"`
	Src         Connect   `xml:"sourceConnector"`
	Dst         []Connect `xml:"destinationConnectors>connector"`
	Name        string    `xml:"name"`
	Description string    `xml:"description"`
	Enabled     bool      `xml:"enabled"`
	//Version          string    `xml:"version"`
	//LastModifiedTime string    `xml:"lastModified>time"`
	//Revision         string    `xml:"revision"`
}

// Connect is used by <sourceConnector> and <destinationConnectors>
type Connect struct {
	Name        string     `xml:"name"`
	Props       []Property `xml:"properties>property"`
	ProtocolIn  string     `xml:"transformer>inboundProtocol"`
	ProtocolOut string     `xml:"transformer>outboundProtocol"`
}

// Disabled returns the string "Disabled" if Channel.Enabled == false
func (c Channel) Disabled() string {
	if !c.Enabled {
		return "Disabled"
	}
	return ""
}

// Property represents each properties of a Connect
type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",innerxml"`
}

const lineSeparator, delimiter, multipleValues = "\r\n", ",", "; "

func main() {
	//Command line flags.
	xmlDir := flag.String("xmlDir", ".", "Directory to parse exported XML Mirth channel files.")
	flag.Parse()

	log.SetPrefix("ERROR: ")
	log.SetFlags(log.Lshortfile)

	output := []byte(strings.Join([]string{
		"Name",
		"Description",
		"Source Data Type",
		"Source Protocol : Address",
		"Destination Data Type",
		"Destination Protocol : Address" + lineSeparator,
	}, delimiter))

	//Process each file in specified xmlDir directory.
	err := filepath.Walk(*xmlDir, func(path string, details os.FileInfo, err error) error {
		if details != nil && !details.IsDir() && strings.HasSuffix(strings.ToLower(path), ".xml") {
			output = append(output, processXMLFile(path)...)
		}
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprintf(os.Stdout, "%s", output)
}

func processXMLFile(path string) []byte {
	src, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln(err, path)
	}

	var c Channel

	//Parse XML data into a Channel struct.
	err = xml.Unmarshal(src, &c)
	if err != nil {
		log.Fatalln(err, path)
	}

	c.Src.ProtocolIn = hl7Version(c.Src.ProtocolIn)

	//Channels can have multiple destinations, so assemble all their properties.
	var destinations, dstProtocols []string
	for _, s := range c.Dst {
		destinations = append(destinations, printSource(s.Props, path))
		dstProtocols = append(dstProtocols, hl7Version(s.ProtocolOut))
	}

	list := []string{
		c.Disabled(),
		strings.TrimSpace(c.Name),
		replaceNewLines(c.Description),
		c.Src.ProtocolIn,
		printSource(c.Src.Props, path),
		strings.Join(dstProtocols, multipleValues),
		strings.Join(destinations, multipleValues),
	}

	return []byte(strings.Join(list, delimiter) + lineSeparator)
}

// printSource determines which function to call based on the property's DataType value.
// Mirth uses the same XML data structure <property name="DataType">Value</property> for all connection types; otherwise this function wouldn't be required.
func printSource(p []Property, path string) string {
	for _, ty := range p {
		if ty.Name != "DataType" {
			continue
		}
		switch ty.Value {
		case "File Reader":
			return fileReader(p)
		case "File Writer":
			return fileWriter(p)
		case "Channel Reader", "Channel Writer":
			return ty.Value
		case "Database Writer":
			return dbWriter(p)
		case "JavaScript Reader", "JavaScript Writer":
			return jsWriter(p)
		case "LLP Listener", "LLP Sender":
			return llpListener(p)
		case "SMTP Sender":
			return smtpSender(p)
		case "HTTP Sender":
			return httpSender(p)
		case "Email Sender":
			return emailSender(p)
		case "HTTP Listener":
			return httpListener(p)
		case "Web Service Sender":
			return webService(p)
		case "Document Writer":
			return docWriter(p)
		default:
			log.Fatalf("%v not defined: %v", ty.Value, path)
		}
	}
	return ""
}

func fileReader(properties []Property) (host string) {
	for _, p := range properties {
		if p.Name == "host" {
			host = p.Value
			break
		}
	}
	return fmt.Sprintf("FILE: %v", host)
}

func fileWriter(properties []Property) (host string) {
	for _, p := range properties {
		if p.Name == "host" {
			host = p.Value
			break
		}
	}
	return fmt.Sprintf("FTP: %v", host)
}

func dbWriter(properties []Property) string {
	for _, p := range properties {
		if p.Name == "URL" && p.Value != "" {
			return p.Value
		}
	}
	return "DB:"
}

func llpListener(properties []Property) string {
	var host, port, template string
	for _, p := range properties {
		if p.Name == "host" {
			host = p.Value
		} else if p.Name == "port" {
			port = p.Value
		} else if p.Name == "template" {
			template = p.Value
		}
		if host != "" && port != "" && template != "" {
			break
		}
	}
	return fmt.Sprintf("LLP: %v:%v/%v", host, port, template)
}

func smtpSender(properties []Property) string {
	var host, port string
	for _, p := range properties {
		if p.Name == "smtpHost" {
			host = p.Value
		} else if p.Name == "smtpPort" {
			port = p.Value
		}
		if host != "" && port != "" {
			break
		}
	}
	return fmt.Sprintf("SMTP: %v:%v", host, port)
}

func httpSender(properties []Property) string {
	for _, p := range properties {
		if p.Name == "host" {
			return p.Value
		}
	}
	return "HTTP:"
}

func httpListener(properties []Property) string {
	var host, port string
	for _, p := range properties {
		if p.Name == "host" {
			host = p.Value
		} else if p.Name == "port" {
			port = p.Value
		}
		if host != "" && port != "" {
			break
		}
	}
	return fmt.Sprintf("HTTP://%v:%v", host, port)
}

func emailSender(properties []Property) string {
	var host, port, from, subject string
	for _, p := range properties {
		if p.Name == "hostname" {
			host = p.Value
		} else if p.Name == "smtpPort" {
			port = p.Value
		} else if p.Name == "fromAddress" {
			from = p.Value
		} else if p.Name == "subject" {
			port = p.Value
		}
		if host != "" && port != "" {
			break
		}
	}
	return fmt.Sprintf("SMTP: %v:%v/%v>%v", host, port, from, subject)
}

func jsWriter(properties []Property) (host string) {
	for _, p := range properties {
		if p.Name == "host" {
			host = p.Value
			break
		}
	}
	return fmt.Sprintf("JS: %v", host)
}

func webService(properties []Property) (url string) {
	for _, p := range properties {
		if p.Name == "dispatcherWsdlUrl" {
			url = p.Value
			break
		}
	}
	return fmt.Sprintf("SOAP: %v", url)
}

func docWriter(properties []Property) (typ string) {
	var host, pattern, docType string
	for _, p := range properties {
		if p.Name == "host" {
			host = p.Value
		} else if p.Name == "outputPattern" {
			pattern = p.Value
		} else if p.Name == "documentType" {
			docType = strings.ToUpper(p.Value)
		}
		if host != "" && pattern != "" && docType != "" {
			break
		}
	}
	return fmt.Sprintf("%v: %v/%v", docType, host, pattern)
}

func hl7Version(h string) string {
	if h == "HL7V2" {
		return "HL7 2.x"
	}
	return h
}

func replaceNewLines(i string) string {
	i = strings.TrimSpace(i)
	i = strings.Replace(i, "\r\n", ". ", -1)
	i = strings.Replace(i, "\n", ". ", -1)
	i = strings.Replace(i, "\r", "", -1)
	i = strings.Replace(i, ",", ";", -1)
	return i
}
