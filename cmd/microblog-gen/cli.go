package main

import (
	"flag"
	"log"
)

func main() {
	sourceDirectory := flag.String("i", "./src", "Source directory with HTML template and other assets.")
	blogDirectory := flag.String("b", "./blog", "Directory that contains blog posts as Markdown files.")
	outputDirectory := flag.String("o", "./build", "Output directory for generated files.")
	force := flag.Bool("f", false, "Overwrite output directory contents.")
	templateFile := flag.String("t", "", "Path to a HTML template for generated blog posts")

	flag.Parse()

	if err := build(*sourceDirectory, *blogDirectory, *outputDirectory, buildOptions{
		Force:            *force,
		PostTemplateFile: *templateFile,
	}); err != nil {
		log.Fatal(err)
	}
}
