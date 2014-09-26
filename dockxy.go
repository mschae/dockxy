package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

var (
	dockerIP     = flag.String("dockerIP", "192.168.59.103", "The IP address of the docker host")
	dockerURL    = flag.String("dockerURL", "tcp://192.168.59.103:2375", "The address to docker (e.g. tcp://10.0.0.1:1234)")
	templatePath = flag.String("templatePath", "templates/site.tmpl", "Path to the nginx template")
	outDir       = flag.String("outDir", "out", "Directory for the generated config files")
)

type Container struct {
	Command string
	Created int64
	Id      string
	Image   string
	Names   []string
	Ports   []map[string]interface{}
	Status  string
}

type Config struct {
	Name string
	Port float64
	IP   string
}

func regenerateConfigFiles() {
	os.RemoveAll(*outDir)
	os.Mkdir(*outDir, 0777)

	fetchContainers()

	exec.Command("nginx", "-s", "reload").Output()
}

func fetchContainers() {
	dockerHTTPURL := strings.Replace(*dockerURL, "tcp", "http", 1)
	resp, _ := http.Get(dockerHTTPURL + "/containers/json?all=1")
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var containers []Container
	json.Unmarshal(body, &containers)

	for _, container := range containers {
		for _, name := range container.Names {
			real_name := strings.Replace(name, "/", "", 1)
			if !strings.Contains(real_name, "/") {
				fmt.Println(real_name)
				for _, item := range container.Ports {
					for key, value := range item {
						if key == "PublicPort" {
							generateTemplate(Config{real_name, value.(float64), *dockerIP})
						}
						fmt.Printf("  %s: %s\n", key, value)
					}
				}
			}
		}
	}
}

func generateTemplate(config Config) {
	file, err := os.Create(*outDir + "/" + config.Name + ".conf")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	t := template.Must(template.ParseFiles(*templatePath))

	err = t.Execute(writer, config)
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	regenerateConfigFiles()

	dockerout := exec.Command("docker", "-H="+*dockerURL, "events")
	out, err := dockerout.StdoutPipe()
	if err != nil {
		panic(err)
	}

	err = dockerout.Start()
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(out)

	for {
		reader.ReadString('\n')
		regenerateConfigFiles()
	}
}
