package main

import "fmt"
import "log"
import "io/ioutil"
import "strings"
import "text/template"
import "os"

var NGINX_TEMPLATES = map[string]string{
	"laravel": "laravel.template",
	"default": "default.template",
}

type Host struct {
	ServerName []string
	Directory  string
	Template   string
}

func (h *Host) DefaultHost() string {
	if len(h.ServerName) == 0 {
		log.Fatal("ServerName is empty: ", h)
	}
	return h.ServerName[0]
}

func (h *Host) Save() {
	tmpl := NGINX_TEMPLATES[h.Template]

	t := template.New(tmpl)
	t, err := t.Parse(getFileContent(tmpl))
	if err != nil {
		log.Fatal(err)
	}

	fileName := fmt.Sprintf("build/%s.conf", h.DefaultHost())
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = t.Execute(f, h)
	if err != nil {
		fmt.Println(err)
	}
}

func (h *Host) ExtractServerName(line string) {
	valid := strings.Contains(line, "ServerName") || strings.Contains(line, "ServerAlias")
	if valid && line[0] != '#' {
		line = strings.TrimLeft(line, "ServerName")
		line = strings.TrimLeft(line, "ServerAlias")
		line = strings.TrimSpace(line)
		h.ServerName = append(h.ServerName, line)
	}
}

func (h *Host) ExtractDirectory(line string) {
	valid := strings.Contains(line, "DocumentRoot")
	if valid && line[0] != '#' {
		line = strings.TrimLeft(line, "DocumentRoot")
		line = strings.TrimSpace(line)
		h.Directory = line

		h.ExtractTemplate()
	}
}

func (h *Host) ExtractTemplate() {
	h.Template = "default"
	if strings.Contains(h.Directory, "/public") {
		h.Template = "laravel"
	}
}

type Block struct {
	First   int
	Last    int
	Counter int
	Data    []string
}

func (b *Block) Valid() bool {
	if b.First > -1 && b.Last > -1 {
		return true
	}
	return false
}
func (b *Block) Reset() {
	b.First = -1
	b.Last = -1
}

func (b *Block) Capture(data []string) {
	b.Data = data[b.First+1 : b.Last]
}

func (b *Block) Increment() {
	b.Counter++
}

func (b *Block) CreateHost() *Host {
	h := new(Host)
	for i := 0; i < len(b.Data); i++ {
		line := strings.TrimSpace(b.Data[i])
		h.ExtractServerName(line)
		h.ExtractDirectory(line)
	}
	return h
}

func (b *Block) CaptureOpenTagIndex(data string, index int) {
	if strings.Contains(data, "<VirtualHost") {
		b.First = index
	}
}

func (b *Block) CaptureCloseTagIndex(data string, index int) {
	if strings.Contains(data, "</VirtualHost>") {
		b.Last = index
	}
}

type Conf struct {
	Content []string
}

func (c *Conf) Read(path string) {
	c.Content = strings.Split(getFileContent(path), "\n")
}

func (c *Conf) Convert() {
	b := Block{First: -1, Last: -1, Counter: 0}

	for i := 0; i < len(c.Content); i++ {
		b.CaptureOpenTagIndex(c.Content[i], i)
		b.CaptureCloseTagIndex(c.Content[i], i)

		if b.Valid() {
			b.Capture(c.Content)
			h := b.CreateHost()
			h.Save()
			b.Reset()
			b.Increment()
			fmt.Printf("File created (#%d): %s (%s)\n", b.Counter, h.DefaultHost(), h.Template)
		}
	}
}

func getFileContent(fileName string) string {
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	return string(contents)
}

func main() {
	conf := new(Conf)
	conf.Read("virtual.conf")
	conf.Convert()
}
