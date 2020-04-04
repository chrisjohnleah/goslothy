/*
Copyright © 2020 Chris Leah <chrisjohnleah@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"fmt"
	Aurora "github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"log"
	"os"
	"regexp"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var filePath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goslothy",
	Short: "goSlothy is slow_error log analyser.",
	Long: `goSlothy is a simple Go program based off of Josh Collinsworth's node version, designed to read a slow_error log files from a WordPress site, 
count the errors and sort the number of slow_errors by their sources.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if filePath == "" {
			fmt.Println(Aurora.Gray(1-1, "Error:").BgRed(), "I need a WordPress slow log file to work with.")
			fmt.Println("Try it like this: slothy path/to/log")
            os.Exit(1)
		}

		fmt.Println(Aurora.Yellow("Welcome to goSlothy v"), Aurora.Yellow("- the faster WordPress slowlog analyser!"))
		fmt.Println("(Only use this tool on valid WP slowlog files, or weird stuff will happen.)")

		fmt.Println(Aurora.Blue("Getting file"))

		lines := make(chan string)
		readerr := make(chan error)

		errorsToPush := make(map[string]int)
		allErrorSources := make(map[string]int)
		totalErrors := 0
		thisError := make([]string, 0)

		go GetLine(filePath, lines, readerr)
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
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goslothy.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.Flags().StringVarP(&filePath, "filepath", "f", "", "Enter the file path to the slow_error log")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".goslothy" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".goslothy")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
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
