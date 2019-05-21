package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	Aurora "github.com/logrusorgru/aurora"
)

func main() {
	version := "1.0.0"

	filepath := flag.String("filepath", "", "The file to parse")
	flag.Parse()

	if *filepath == "" {
		fmt.Println(Aurora.Gray(1-1, "Error:").BgRed(), "I need a WordPress slow log file to work with.")
		fmt.Println("Try it like this: slothy path/to/log")
		return
	}

	fmt.Println(Aurora.Yellow("Welcome to goSlothy v"), Aurora.Yellow(version), Aurora.Yellow("- the faster WordPress slowlog analyzer!"))
	fmt.Println("(Only use this tool on valid WP slowlog files, or weird stuff will happen.)")

	fmt.Println(Aurora.Blue("Getting file"))

	lines := make(chan string)
	readerr := make(chan error)

	errorsToPush := make(map[string]int)
	allErrorSources := make(map[string]int)
	totalErrors := 0
	thisError := make([]string, 0)

	go GetLine(*filepath, lines, readerr)
	countCore := 0
	countPluginsThemes := 0
	for line := range lines {

		matchWpCoreFiles, _ := regexp.MatchString("^$", line)
		matchPluginsThemes, _ := regexp.MatchString("/(plugins|themes)/[^/]*/i", line)

		if matchWpCoreFiles {

			if len(errorsToPush) == 0 {
				errorsToPush["no-plugin-or-theme"] = 1
				// fmt.Println(errorsToPush["**wp-core**"])
			}

			for errorToPush := range errorsToPush {
				if allErrorSources[errorToPush] > 0 {
					allErrorSources[errorToPush]++
					// fmt.Println(allErrorSources[errorToPush])
				} else {
					allErrorSources[errorToPush] = 1
					// fmt.Println(allErrorSources[errorToPush])
				}
			}

			thisError = make([]string, 0)
			totalErrors++

			countCore++
			// fmt.Println(line)
		} else if matchPluginsThemes {
			countPluginsThemes++
			themeAndPluginErrors := append(thisError, line)
			fmt.Printf("%v", themeAndPluginErrors)
		}

	}

	if err := <-readerr; err != nil {
		log.Fatal(err)
	}

	fmt.Println(Aurora.Green("Processing Complete"), countCore, countPluginsThemes)
	fmt.Printf("%v", allErrorSources)

}
func GetLine(filepath string, lines chan string, readerr chan error) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines <- scanner.Text()
	}
	close(lines) // close causes range on channel to break out of loop
	readerr <- scanner.Err()
}
