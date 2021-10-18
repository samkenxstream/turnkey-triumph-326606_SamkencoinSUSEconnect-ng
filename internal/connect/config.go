package connect

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	defaultConfigPath = "/etc/SUSEConnect"
	defaultBaseURL    = "https://scc.suse.com"
	defaultInsecure   = false
)

// Config holds the config!
type Config struct {
	Path             string
	BaseURL          string `json:"url"`
	Language         string `json:"language"`
	Insecure         bool   `json:"insecure"`
	Namespace        string `json:"namespace"`
	FsRoot           string
	Token            string
	Product          Product
	InstanceDataFile string
	Email            string `json:"email"`
	NoZypperRefresh  bool
}

// NewConfig returns a Config with defaults
func NewConfig() Config {
	return Config{
		Path:     defaultConfigPath,
		BaseURL:  defaultBaseURL,
		Insecure: defaultInsecure,
	}
}

func (c Config) toYAML() []byte {
	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, "---\n")
	fmt.Fprintf(&buf, "url: %s\n", c.BaseURL)
	fmt.Fprintf(&buf, "insecure: %v\n", c.Insecure)
	if c.Language != "" {
		fmt.Fprintf(&buf, "language: %s\n", c.Language)
	}
	if c.Namespace != "" {
		fmt.Fprintf(&buf, "namespace: %s\n", c.Namespace)
	}
	return buf.Bytes()
}

// Save saves the config to Path
func (c Config) Save() error {
	data := c.toYAML()
	return os.WriteFile(c.Path, data, 0644)
}

// Load tries to read and merge the settings from Path.
// Ignore errors as it's quite normal that Path does not exist.
func (c *Config) Load() {
	f, err := os.Open(c.Path)
	if err != nil {
		Debug.Println(err)
		return
	}
	defer f.Close()
	parseConfig(f, c)
	Debug.Printf("Config after parsing: %+v", c)
}

func parseConfig(r io.Reader, c *Config) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if strings.HasPrefix(key, "#") {
			continue
		}
		switch key {
		case "url":
			c.BaseURL = val
		case "language":
			c.Language = val
		case "namespace":
			c.Namespace = val
		case "insecure":
			c.Insecure, _ = strconv.ParseBool(val)
		case "no_zypper_refs":
			c.NoZypperRefresh, _ = strconv.ParseBool(val)
		default:
			Debug.Printf("Cannot parse line \"%s\" from %s", line, c.Path)
		}
	}
}

// MergeJSON merges attributes of jsn that match Config fields
func (c *Config) MergeJSON(jsn string) error {
	return json.Unmarshal([]byte(jsn), c)
}
