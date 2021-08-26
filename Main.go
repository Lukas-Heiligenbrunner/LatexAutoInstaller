package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
)

func main() {
	fmt.Printf("Pdflatex command exists: %t\n", commandExists("pdflatex"))
	fmt.Printf("%s/%s\n", runtime.GOOS, runtime.GOARCH)

	compileAndInstall()
}

func compileAndInstall() {
	out, err := compileLatex("main.tex")
	//fmt.Println(*out)
	if err != nil {
		fmt.Println("An error occured while compiling the document!")

		filename := parseMissingFile(out)
		if filename != "" {
			fmt.Printf("We need to download: %s\n", filename)

			// now we neet to perform a root check
			if rootCheck() {
				log.Println("Awesome! You are now running this program with root permissions!")

				if installFile(filename) {
					// if successfully installed we try to compile again
					compileAndInstall()
				}
			} else {
				log.Fatal("This program must be run as root! (sudo)")
			}
		} else {
			fmt.Println(*out)

			fmt.Println("another build error occured!")
		}
	} else {
		fmt.Println("document built successfully!")
	}
}

func parseMissingFile(output *string) string {
	matchfile := regexp.MustCompile("! LaTeX Error: File `([^`']*)' not found|! I can't find file `([^`']*)'.")
	matches := matchfile.FindStringSubmatch(*output)
	fmt.Printf("%#v\n", matches)
	if matches != nil {
		if matches[1] != "" {
			return matches[1]
		} else {
			return matches[2]
		}
	}

	// ok now we try to find a font error
	fontregex := regexp.MustCompile(`! Font \\[^=]*=([^\s]*)\s`)
	fontmatch := fontregex.FindStringSubmatch(*output)
	fmt.Printf("%#v\n", fontmatch)
	if fontmatch != nil {
		if fontmatch[1] != "" {
			return fontmatch[1]
		}
	}

	// now try babel errors
	babelregex := regexp.MustCompile("Unknown option `([^`']*)'. Either you misspelled")
	babelmatch := babelregex.FindStringSubmatch(*output)
	if babelmatch != nil {
		if babelmatch[1] != "" {
			return babelmatch[1] + ".ldf"
		}
	}

	return ""
}

// check if a specific system command is available
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// parse the thumbail picture from video file
func compileLatex(filename string) (*string, error) {
	app := "pdflatex"

	cmd := exec.Command(app,
		"-file-line-error",
		"-interaction=nonstopmode",
		"-synctex=1",
		"-output-format=pdf",
		filename)

	stdout, err := cmd.Output()

	output := string(stdout)

	return &output, err
}

func rootCheck() bool {
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	// output has trailing \n
	// need to remove the \n
	// otherwise it will cause error for strconv.Atoi
	// log.Println(output[:len(output)-1])

	// 0 = root, 501 = non-root user
	i, err := strconv.Atoi(string(output[:len(output)-1]))

	if err != nil {
		log.Fatal(err)
	}

	if i == 0 {
		return true
	} else {
		return false
	}
}

func installFile(filename string) bool {
	if !commandExists("dnf") {
		fmt.Println("dnf not existing!")
		return false
	}

	cmd := exec.Command("dnf", "-y", "install", fmt.Sprintf("tex(%s)", filename))
	fmt.Println(cmd.String())

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	fmt.Println("running dnf install now!")

	go func() {
		merged := io.MultiReader(stderr, stdout)
		scanner := bufio.NewScanner(merged)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
	}()

	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
